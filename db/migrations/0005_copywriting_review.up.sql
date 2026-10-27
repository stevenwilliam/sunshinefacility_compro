-- 0005 — copywriting review of 2026-10-27 (SFG_Website_Copywriting_Review_1.xlsx,
-- tab "Yang ini"): the rows not yet done in the CMS, plus the gaps left by the
-- rows that were (rows 5 and 30, and the Indonesian of the CMS edits).
-- Every UPDATE is guarded by the text it replaces, so a row edited in the CMS
-- since keeps its edit. Content edits clear the placeholder flag (BR-2).

-- ---------------------------------------------------------------- site_text

-- Row 2: hero eyebrow.
UPDATE site_text SET value_en = 'Integrated Solutions for Business Operations',
                     value_id = 'Solusi Terpadu untuk Operasional Bisnis', is_placeholder = false, updated_at = now()
 WHERE key = 'home.hero.eyebrow' AND value_en = 'Integrated facility services · Indonesia';

-- Row 7: trust line under the hero.
UPDATE site_text SET value_en = 'Food safety and hygiene standards you can trust.',
                     value_id = 'Standar keamanan pangan dan higiene yang dapat Anda andalkan.', is_placeholder = false, updated_at = now()
 WHERE key = 'home.hero.cert_line' AND value_en = 'HACCP · ISO 22000:2018 · SLHS certified';

-- Row 17: the same question on Home and on Services.
UPDATE site_text SET value_en = 'Not Sure What Your Operations Need?',
                     value_id = 'Belum yakin apa yang dibutuhkan operasional Anda?', is_placeholder = false, updated_at = now()
 WHERE key = 'home.services.unsure' AND value_en = 'Not sure which services your building needs?';

-- Row 31: WhatsApp label.
UPDATE site_text SET value_en = 'WhatsApp Our Team',
                     value_id = 'Hubungi Tim Kami via WhatsApp', is_placeholder = false, updated_at = now()
 WHERE key = 'wa.chat' AND value_en = 'Chat on WhatsApp';

-- Row 5 (yellow, not yet live): the quotation buttons.
UPDATE site_text SET value_en = 'Discuss Your Requirements',
                     value_id = 'Diskusikan Kebutuhan Anda', is_placeholder = false, updated_at = now()
 WHERE key IN ('cta.quote', 'home.ball.quote') AND lower(value_en) = 'request a quotation';

-- Indonesian of the yellow rows edited in English only.
UPDATE site_text SET value_id = 'Jelajahi Solusi Kami', updated_at = now()
 WHERE key = 'home.hero.cta_services' AND value_en = 'Explore Our Solutions' AND value_id = 'Layanan Kami';
UPDATE site_text SET value_id = 'Beragam solusi terpadu.', updated_at = now()
 WHERE key = 'home.hero.title_accent' AND value_en = 'Multiple integrated solutions.' AND value_id = 'Setiap shift, tercatat.';

-- ------------------------------------------------------------ page_sections

-- Indonesian of the yellow rows 3–4 (home hero).
UPDATE page_sections SET title_id = 'Satu mitra tepercaya.',
       body_id = 'Sunshine Facilities Group menyatukan layanan-layanan penting, mulai dari layanan makanan dan manajemen fasilitas hingga solusi tenaga kerja dan rantai pasok, dalam satu mitra yang bertanggung jawab penuh.',
       updated_at = now()
 WHERE page = 'home' AND section = 'hero' AND title_en = 'One trusted partner.' AND title_id = 'Gedung bersih dan terkelola.';

-- Indonesian of the yellow rows 12–13 (home quote).
UPDATE page_sections SET title_id = 'Ceritakan kebutuhan bisnis Anda.',
       body_id = 'Sampaikan prioritas operasional Anda kepada kami. Kami akan menghubungkan Anda dengan tim yang tepat untuk menemukan solusi yang sesuai.',
       updated_at = now()
 WHERE page = 'home' AND section = 'quote' AND title_en = 'Tell Us What Your Business Needs' AND title_id = 'Ceritakan tentang gedung Anda.';

-- Row 14: Services page title.
UPDATE page_sections SET title_en = 'Integrated Services for the Way Your Business Operates',
       title_id = 'Layanan terpadu yang mengikuti cara bisnis Anda beroperasi', is_placeholder = false, updated_at = now()
 WHERE page = 'services' AND section = 'hero' AND title_en = 'Five business lines. One partner on site.';

-- Row 17: Services closing question.
UPDATE page_sections SET title_en = 'Not Sure What Your Operations Need?',
       title_id = 'Belum yakin apa yang dibutuhkan operasional Anda?', is_placeholder = false, updated_at = now()
 WHERE page = 'services' AND section = 'cta' AND title_en = 'Not sure which services your building needs?';

-- Rows 18–19: Industries title and introduction.
UPDATE page_sections SET title_en = 'Solutions for the Needs of Every Industry We Serve.',
       title_id = 'Solusi yang menjawab kebutuhan setiap industri yang kami layani.',
       body_en = 'Every industry has its own operational priorities, service standards and site requirements. SFG adapts its teams and service approach to the needs of each environment.',
       body_id = 'Setiap industri punya prioritas operasional, standar layanan, dan kebutuhan lokasi yang berbeda. SFG menyesuaikan tim dan pendekatan layanannya dengan kebutuhan masing-masing lingkungan.',
       is_placeholder = false, updated_at = now()
 WHERE page = 'industries' AND section = 'hero' AND title_en = 'Built around how each kind of building runs.';

-- Row 22: vision (was marked draft).
UPDATE page_sections SET body_en = 'To be a trusted integrated solutions partner that enables businesses to focus on their core operations.',
       body_id = 'Menjadi mitra solusi terpadu tepercaya yang memungkinkan setiap bisnis fokus pada operasional intinya.',
       is_placeholder = false, updated_at = now()
 WHERE page = 'about' AND section = 'vision' AND body_en = 'To be Indonesia’s most trusted integrated facility solutions partner.';

-- The vision and mission are now confirmed wording: drop the section's draft tag.
UPDATE page_sections SET is_placeholder = false, updated_at = now()
 WHERE page = 'about' AND section = 'vision_mission' AND is_placeholder;

-- Row 27: purpose, as the opening paragraph of the company story.
UPDATE page_sections SET body_en = 'We enable our clients to focus on their core business by taking on the operational services they rely on every day.' || E'\n\n' || body_en,
       body_id = 'Kami membantu klien fokus pada bisnis inti mereka dengan mengambil alih layanan operasional yang mereka andalkan setiap hari.' || E'\n\n' || body_id,
       is_placeholder = false, updated_at = now()
 WHERE page = 'about' AND section = 'story' AND body_en LIKE 'Building on this foundation,%';

-- Row 28: Careers title.
UPDATE page_sections SET title_en = 'Join the teams that keep essential operations running.',
       title_id = 'Bergabunglah dengan tim yang menjaga operasional penting tetap berjalan.', is_placeholder = false, updated_at = now()
 WHERE page = 'careers' AND section = 'hero' AND title_en = 'Your operations. Our people.';

-- Rows 29–30: Contact title and introduction.
UPDATE page_sections SET title_en = 'Tell Us What Your Business Needs.',
       title_id = 'Ceritakan kebutuhan bisnis Anda.',
       body_en = 'Share your requirements with us. Our team will review your enquiry and contact you to discuss the next steps.',
       body_id = 'Sampaikan kebutuhan Anda kepada kami. Tim kami akan meninjau permintaan Anda dan menghubungi Anda untuk membahas langkah selanjutnya.',
       is_placeholder = false, updated_at = now()
 WHERE page = 'contact' AND section = 'hero' AND title_en IN ('Discuss Your Requirements', 'Request a quotation.')
   AND body_en = 'Tell us about your building. We will arrange a site survey and send an itemised proposal, usually within one working day.';

-- ------------------------------------------------------------ content_items

-- Rows 10–11: "Why Sunshine" titles; the text under them stays.
UPDATE content_items SET title_en = 'Transparent Operations, Documented Delivery',
       title_id = 'Operasional transparan, pekerjaan terdokumentasi', is_placeholder = false, updated_at = now()
 WHERE list_key = 'home_why' AND title_en = 'Digital reporting & proof of work';
UPDATE content_items SET title_en = 'A Dedicated Point of Contact',
       title_id = 'Narahubung khusus untuk Anda', is_placeholder = false, updated_at = now()
 WHERE list_key = 'home_why' AND title_en = '24/7 response';

-- Row 26: the recommended mission set replaces rows 23–25.
UPDATE content_items SET body_en = 'Enable clients to focus on their core business through integrated service solutions.',
       body_id = 'Memungkinkan klien fokus pada bisnis inti melalui solusi layanan terpadu.', is_placeholder = false, updated_at = now()
 WHERE list_key = 'about_mission' AND body_en = 'Deliver safe, consistent services through trained and supervised teams.';
UPDATE content_items SET body_en = 'Deliver safe, reliable and consistent services through capable, accountable teams.',
       body_id = 'Menghadirkan layanan yang aman, andal, dan konsisten melalui tim yang cakap dan bertanggung jawab.', is_placeholder = false, updated_at = now()
 WHERE list_key = 'about_mission' AND body_en = 'Make operations transparent, with a digital record of every job.';
UPDATE content_items SET body_en = 'Simplify operations through coordinated delivery, clear accountability and transparent reporting.',
       body_id = 'Menyederhanakan operasional melalui pelaksanaan yang terkoordinasi, tanggung jawab yang jelas, dan pelaporan yang transparan.', is_placeholder = false, updated_at = now()
 WHERE list_key = 'about_mission' AND body_en = 'Simplify our clients’ operations with one partner for food, facilities, workforce and supply.';

-- --------------------------------------------------------------- industries

-- Rows 20–21: each industry's own needs first, then the SFG services that
-- meet them (the lines mapped to it in industry_lines).
UPDATE industries SET body_en = 'Supporting healthcare environments. Hospitals and clinics need strict hygiene, clinical nutrition and round-the-clock reliability. We combine patient and staff food services, infection-aware cleaning, trained support staff and digital service records.',
       body_id = 'Mendukung lingkungan layanan kesehatan. Rumah sakit dan klinik membutuhkan standar higiene yang ketat, gizi klinis, dan layanan yang andal setiap saat. Kami memadukan layanan makanan untuk pasien dan staf, kebersihan berstandar pengendalian infeksi, tenaga pendukung terlatih, serta pencatatan layanan secara digital.',
       updated_at = now()
 WHERE key = 'healthcare' AND body_en LIKE 'Clinical nutrition and hospital food operations,%';
UPDATE industries SET body_en = 'Offices need well-run workplaces that support productivity and a good employee experience. We combine canteen and pantry services, cleaning and building upkeep, support staff and digital reporting.',
       body_id = 'Perkantoran membutuhkan lingkungan kerja yang tertata untuk menunjang produktivitas dan kenyamanan karyawan. Kami memadukan layanan kantin dan pantry, kebersihan dan perawatan gedung, tenaga pendukung, serta pelaporan digital.',
       updated_at = now()
 WHERE key = 'corporate' AND body_en LIKE 'Canteens, pantries and coffee,%';
UPDATE industries SET body_en = 'Malls and stores need clean, safe spaces that stay presentable through long trading hours and heavy footfall. We combine large-area cleaning, security and support staff with food solutions for tenants and visitors.',
       body_id = 'Mal dan toko membutuhkan area yang bersih, aman, dan tetap rapi sepanjang jam operasional yang panjang dengan arus pengunjung yang tinggi. Kami memadukan kebersihan area luas, keamanan, dan tenaga pendukung dengan solusi makanan untuk tenant dan pengunjung.',
       updated_at = now()
 WHERE key = 'retail' AND body_en LIKE 'Large-scale retail operational support:%';
UPDATE industries SET body_en = 'Plants, mines and remote camps need high-volume meals, dependable supply and teams that can work far from cities. We combine large-scale canteens, remote supply and logistics, an on-site workforce and site maintenance.',
       body_id = 'Pabrik, tambang, dan camp terpencil membutuhkan makanan dalam jumlah besar, pasokan yang andal, dan tim yang siap bekerja jauh dari kota. Kami memadukan kantin berkapasitas besar, pasokan dan logistik ke lokasi terpencil, tenaga kerja di lokasi, serta perawatan site.',
       updated_at = now()
 WHERE key = 'industrial' AND body_en LIKE 'High-volume canteens, supply chain and site services%';
UPDATE industries SET body_en = 'Hotels, restaurants and food brands need consistent quality that guests notice and standards that hold every day. We combine food operations from concept to execution, trained service staff and facility upkeep.',
       body_id = 'Hotel, restoran, dan merek makanan membutuhkan kualitas yang konsisten dan terasa oleh tamu, dengan standar yang terjaga setiap hari. Kami memadukan operasional makanan dari konsep hingga eksekusi, staf layanan terlatih, serta perawatan fasilitas.',
       updated_at = now()
 WHERE key = 'hospitality' AND body_en LIKE 'Quality dining and operational services%';
UPDATE industries SET body_en = 'Schools and campuses need healthy meals, clean classrooms and a safe environment for students and staff. We combine nutritious food services with cleaning, maintenance and security.',
       body_id = 'Sekolah dan kampus membutuhkan makanan sehat, ruang kelas yang bersih, dan lingkungan yang aman bagi siswa dan staf. Kami memadukan layanan makanan bergizi dengan kebersihan, perawatan, dan keamanan.',
       updated_at = now()
 WHERE key = 'education' AND body_en LIKE 'Meals, cleaning and safety for schools%';
UPDATE industries SET body_en = 'Residential towers need common areas that stay clean, safe and well maintained for residents around the clock. We combine common-area cleaning, MEP upkeep, security and trained building staff.',
       body_id = 'Apartemen dan hunian vertikal membutuhkan area bersama yang selalu bersih, aman, dan terawat bagi penghuni setiap saat. Kami memadukan kebersihan area umum, perawatan MEP, keamanan, dan staf gedung yang terlatih.',
       updated_at = now()
 WHERE key = 'residential' AND body_en LIKE 'Common-area cleaning, MEP upkeep and security%';

-- ------------------------------------------------------------ service_lines

-- Rows 8–9 and 15–16: Remote is how the portfolio reaches remote sites, not
-- a separate service. The short label "Remote" (chips, filters) stays.
UPDATE service_lines SET title_en = 'Remote-Managed, On-Site Delivered',
       title_id = 'Dikelola dari Jauh, Dilaksanakan di Lokasi',
       summary_en = 'From procurement and production to logistics and on-site execution, SFG coordinates the support behind essential services—even in remote operating locations.',
       summary_id = 'Mulai dari pengadaan dan produksi hingga logistik dan pelaksanaan di lokasi, SFG mengoordinasikan dukungan di balik layanan-layanan penting, termasuk di lokasi operasional yang terpencil.',
       description_en = 'We coordinate the people, supplies, logistics and on-site execution needed to support essential operations in remote or complex locations. Service scope is tailored to each site’s requirements.',
       description_id = 'Kami mengoordinasikan tenaga kerja, pasokan, logistik, dan pelaksanaan di lokasi yang dibutuhkan untuk mendukung operasional penting di lokasi terpencil maupun kompleks. Cakupan layanan disesuaikan dengan kebutuhan setiap lokasi.',
       lead_en = 'Support for Remote and Complex Operating Locations',
       lead_id = 'Dukungan untuk lokasi operasional yang terpencil dan kompleks.',
       updated_at = now()
 WHERE key = 'remote' AND title_en = 'Remote';
