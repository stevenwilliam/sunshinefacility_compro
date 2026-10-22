package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/access"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/domain/filetype"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/apierror"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/id"
	"github.com/stevenwilliam/sunshinefacility_compro/internal/platform/sanitize"
)

// Rendition widths per kind (BR-4.2). Never upscaled: an image narrower than
// a width simply does not get that rendition.
var variantWidths = map[string][]int{
	"photo":       {480, 960, 1600},
	"logo":        {240, 480},
	"certificate": {480, 960, 1600, 2400},
	"brand":       {480, 960},
}

var mediaKinds = []string{"photo", "logo", "certificate", "brand"}

type Media struct {
	store    Store
	storage  Storage
	images   ImageProcessor
	settings *Settings
	public   *Public
}

func NewMedia(store Store, storage Storage, images ImageProcessor, settings *Settings, public *Public) *Media {
	return &Media{store: store, storage: storage, images: images, settings: settings, public: public}
}

func (m *Media) List(ctx context.Context, actor Actor, q, kind string) ([]MediaRecord, error) {
	if !access.Can(access.Role(actor.Role), access.ContentRead) {
		return nil, apierror.New(apierror.CodeForbidden, "You do not have access to the media library.")
	}
	return m.store.ListMedia(ctx, q, kind)
}

// Upload — BR-4.1, BR-4.2.
func (m *Media) Upload(ctx context.Context, actor Actor, filename, kind string, data []byte) (*MediaRecord, error) {
	if !access.Can(access.Role(actor.Role), access.MediaWrite) {
		return nil, apierror.New(apierror.CodeForbidden, "You cannot upload media.")
	}
	if !contains(mediaKinds, kind) {
		return nil, apierror.Validation("Choose a media type.", map[string]string{"kind": "Choose Photo, Logo, Certificate or Brand."})
	}
	maxMB := m.settings.Int(ctx, "media.max_upload_mb", 15)
	if int64(len(data)) > int64(maxMB)<<20 {
		return nil, apierror.Newf(apierror.CodeTooLarge, "This image is over %d MB. Please upload a smaller file.", maxMB)
	}
	mime, ok := filetype.Image(data)
	if !ok {
		return nil, apierror.New(apierror.CodeUnsupportedType, "Upload a PNG, JPEG or WebP image. Other formats, including SVG, are not accepted.")
	}
	name := sanitize.Filename(filename)
	rec := MediaRecord{ID: id.NewString(), Kind: kind, Name: name}
	if err := m.processRaster(ctx, &rec, data, mime); err != nil {
		return nil, err
	}
	err := m.store.InTx(ctx, func(s Store) error {
		if err := s.InsertMedia(ctx, rec); err != nil {
			return err
		}
		return audit(ctx, s, actor, "upload", "media", rec.ID, "Uploaded "+rec.Name, nil,
			map[string]any{"name": rec.Name, "kind": rec.Kind, "bytes": rec.Bytes, "width": rec.Width, "height": rec.Height})
	})
	if err != nil {
		m.removeObjects(ctx, rec)
		return nil, err
	}
	return &rec, nil
}

// processRaster re-encodes and stores the renditions plus the original.
func (m *Media) processRaster(ctx context.Context, rec *MediaRecord, data []byte, mime string) error {
	maxDim := m.settings.Int(ctx, "media.max_dimension_px", 8000)
	widths := variantWidths[rec.Kind]
	img, err := m.images.Process(data, widths, maxDim)
	if err != nil {
		return apierror.Wrap(apierror.CodeValidation, "This image could not be read, or it is larger than "+itoa(maxDim)+" px. Try another file.", err)
	}
	sum := sha256.Sum256(data)
	rec.Hash = hex.EncodeToString(sum[:])
	rec.Mime, rec.Width, rec.Height, rec.Bytes = mime, img.Width, img.Height, int64(len(data))
	rec.ObjectKey = fmt.Sprintf("media/%s/original%s", rec.ID, filetype.Extension(mime))
	if err := m.storage.Put(ctx, rec.ObjectKey, bytes.NewReader(data), int64(len(data)), mime); err != nil {
		return err
	}
	rec.Variants = nil
	for _, v := range img.Variants {
		key := fmt.Sprintf("media/%s/%d.webp", rec.ID, v.Width)
		if err := m.storage.Put(ctx, key, bytes.NewReader(v.Bytes), int64(len(v.Bytes)), "image/webp"); err != nil {
			return err
		}
		rec.Variants = append(rec.Variants, MediaVariant{W: v.Width, Key: key, Bytes: int64(len(v.Bytes))})
	}
	return nil
}

func (m *Media) UpdateMeta(ctx context.Context, actor Actor, mediaID, name, altEN, altID string) (*MediaRecord, error) {
	if !access.Can(access.Role(actor.Role), access.MediaWrite) {
		return nil, apierror.New(apierror.CodeForbidden, "You cannot change media.")
	}
	errs := map[string]string{}
	nm, err := sanitize.Text(name, 200)
	if err != nil {
		errs["name"] = "Enter a name (up to 200 characters)."
	}
	ae, err1 := sanitize.Optional(altEN, 300)
	ai, err2 := sanitize.Optional(altID, 300)
	if err1 != nil {
		errs["altEn"] = "Alt text must be 300 characters or fewer."
	}
	if err2 != nil {
		errs["altId"] = "Alt text must be 300 characters or fewer."
	}
	if len(errs) > 0 {
		return nil, apierror.Validation("Please correct the highlighted fields.", errs)
	}
	err = m.store.InTx(ctx, func(s Store) error {
		before, err := s.GetMedia(ctx, mediaID)
		if errors.Is(err, ErrNotFound) {
			return apierror.NotFound("Image")
		}
		if err != nil {
			return err
		}
		if err := s.UpdateMediaMeta(ctx, mediaID, nm, ae, ai); err != nil {
			return err
		}
		return audit(ctx, s, actor, "update", "media", mediaID, "Updated "+nm,
			map[string]any{"name": before.Name, "altEn": before.AltEN, "altId": before.AltID},
			map[string]any{"name": nm, "altEn": ae, "altId": ai})
	})
	if err != nil {
		return nil, err
	}
	m.public.Invalidate()
	return m.store.GetMedia(ctx, mediaID)
}

// Delete — BR-4.3: an image in use cannot be deleted. The foreign keys
// refuse it in the database too; this check exists to say HOW MANY items use
// it, which a constraint error cannot.
func (m *Media) Delete(ctx context.Context, actor Actor, mediaID string) error {
	if !access.Can(access.Role(actor.Role), access.MediaWrite) {
		return apierror.New(apierror.CodeForbidden, "You cannot delete media.")
	}
	var rec *MediaRecord
	err := m.store.InTx(ctx, func(s Store) error {
		var err error
		rec, err = s.GetMedia(ctx, mediaID)
		if errors.Is(err, ErrNotFound) {
			return apierror.NotFound("Image")
		}
		if err != nil {
			return err
		}
		n, err := s.MediaUsage(ctx, mediaID)
		if err != nil {
			return err
		}
		if n > 0 {
			return apierror.Newf(apierror.CodeInUse, "Can’t delete: used in %d item%s. Replace it there first.", n, plural(n))
		}
		if err := s.DeleteMedia(ctx, mediaID); err != nil {
			return err
		}
		return audit(ctx, s, actor, "delete", "media", mediaID, "Deleted "+rec.Name, map[string]any{"name": rec.Name, "kind": rec.Kind}, nil)
	})
	if err != nil {
		return err
	}
	m.removeObjects(ctx, *rec)
	return nil
}

func (m *Media) removeObjects(ctx context.Context, rec MediaRecord) {
	if rec.ObjectKey != "" {
		_ = m.storage.Delete(ctx, rec.ObjectKey)
	}
	for _, v := range rec.Variants {
		_ = m.storage.Delete(ctx, v.Key)
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// Object returns the bytes behind a public media URL. The hash in the URL
// must match the row: a stale URL is a 404, never different bytes under a
// URL that promised to be immutable.
func (m *Media) Object(ctx context.Context, mediaID, hashPrefix string, width int, svg bool) ([]byte, string, error) {
	rec, err := m.store.GetMedia(ctx, mediaID)
	if err != nil || rec.Hash == "" || !strings.HasPrefix(rec.Hash, hashPrefix) || len(hashPrefix) < 8 {
		return nil, "", ErrNotFound
	}
	if svg {
		if rec.Mime != "image/svg+xml" {
			return nil, "", ErrNotFound
		}
		return m.read(ctx, rec.ObjectKey, "image/svg+xml")
	}
	for _, v := range rec.Variants {
		if v.W == width {
			return m.read(ctx, v.Key, "image/webp")
		}
	}
	return nil, "", ErrNotFound
}

func (m *Media) read(ctx context.Context, key, ct string) ([]byte, string, error) {
	rc, _, err := m.storage.Get(ctx, key)
	if err != nil {
		return nil, "", err
	}
	defer rc.Close()
	var b bytes.Buffer
	if _, err := b.ReadFrom(rc); err != nil {
		return nil, "", err
	}
	return b.Bytes(), ct, nil
}

// SeedAssets uploads and processes the files for every seeded media row that
// has none yet. It is idempotent: a row with a hash is skipped.
func (m *Media) SeedAssets(ctx context.Context, assetsDir string, log func(string)) error {
	all, err := m.store.ListMedia(ctx, "", "")
	if err != nil {
		return err
	}
	done := 0
	for _, rec := range all {
		if rec.Slug == "" || rec.Hash != "" {
			continue
		}
		path, ok := seedPath(assetsDir, rec)
		if !ok {
			log("skip " + rec.Slug + ": no file mapping")
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", rec.Slug, err)
		}
		r := rec
		if strings.HasSuffix(path, ".svg") {
			if !looksLikeSVG(data) {
				return fmt.Errorf("%s: not an SVG", path)
			}
			sum := sha256.Sum256(data)
			r.Hash, r.Mime, r.Bytes = hex.EncodeToString(sum[:]), "image/svg+xml", int64(len(data))
			r.ObjectKey = "media/" + r.ID + "/original.svg"
			if err := m.storage.Put(ctx, r.ObjectKey, bytes.NewReader(data), int64(len(data)), "image/svg+xml"); err != nil {
				return err
			}
			r.Variants = []MediaVariant{}
		} else {
			mime, ok := filetype.Image(data)
			if !ok {
				return fmt.Errorf("%s: not a PNG, JPEG or WebP", path)
			}
			if err := m.processRaster(ctx, &r, data, mime); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
		}
		if err := m.store.UpdateMediaFile(ctx, r); err != nil {
			return err
		}
		done++
		log(fmt.Sprintf("seeded %s (%d bytes, %d renditions)", r.Slug, r.Bytes, len(r.Variants)))
	}
	log(fmt.Sprintf("seed-assets: %d file(s) processed", done))
	m.public.Invalidate()
	return nil
}

func seedPath(dir string, rec MediaRecord) (string, bool) {
	parts := strings.SplitN(rec.Slug, "/", 2)
	if len(parts) != 2 {
		return "", false
	}
	switch parts[0] {
	case "photo":
		return filepath.Join(dir, "photos", parts[1]+".png"), true
	case "logo":
		return filepath.Join(dir, "clients", parts[1]+".png"), true
	case "certificate":
		return filepath.Join(dir, "certificates", parts[1]+".png"), true
	case "brand":
		return filepath.Join(dir, "brand", filepath.Base(rec.Name)), true
	}
	return "", false
}

func looksLikeSVG(b []byte) bool {
	s := strings.TrimSpace(string(b[:min(len(b), 512)]))
	if strings.Contains(strings.ToLower(string(b)), "<script") {
		return false
	}
	return strings.HasPrefix(s, "<svg") || strings.HasPrefix(s, "<?xml")
}

func contains(list []string, v string) bool { return slices.Contains(list, v) }
