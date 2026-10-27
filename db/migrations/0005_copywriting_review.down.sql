-- 0005 down — restores the copy 0005 replaced, only where the text is still
-- what 0005 wrote (a later CMS edit is kept).

UPDATE site_text SET value_en = 'Integrated facility services · Indonesia', value_id = 'Layanan fasilitas terpadu · Indonesia'
 WHERE key = 'home.hero.eyebrow' AND value_en = 'Integrated Solutions for Business Operations';
UPDATE site_text SET value_en = 'HACCP · ISO 22000:2018 · SLHS certified', value_id = 'Bersertifikat HACCP · ISO 22000:2018 · SLHS'
 WHERE key = 'home.hero.cert_line' AND value_en = 'Food safety and hygiene standards you can trust.';
UPDATE site_text SET value_en = 'Not sure which services your building needs?', value_id = 'Belum yakin layanan apa yang dibutuhkan gedung Anda?'
 WHERE key = 'home.services.unsure' AND value_en = 'Not Sure What Your Operations Need?';
UPDATE site_text SET value_en = 'Chat on WhatsApp', value_id = 'Chat via WhatsApp'
 WHERE key = 'wa.chat' AND value_en = 'WhatsApp Our Team';
UPDATE site_text SET value_en = 'Request a Quotation', value_id = 'Minta Penawaran'
 WHERE key = 'cta.quote' AND value_en = 'Discuss Your Requirements';
UPDATE site_text SET value_en = 'Request a quotation', value_id = 'Minta penawaran'
 WHERE key = 'home.ball.quote' AND value_en = 'Discuss Your Requirements';
UPDATE site_text SET value_id = 'Layanan Kami'
 WHERE key = 'home.hero.cta_services' AND value_id = 'Jelajahi Solusi Kami';
UPDATE site_text SET value_id = 'Setiap shift, tercatat.'
 WHERE key = 'home.hero.title_accent' AND value_id = 'Beragam solusi terpadu.';

UPDATE page_sections SET title_id = 'Gedung bersih dan terkelola.',
       body_id = 'Sunshine Facilities Group menangani layanan makanan, kebersihan, perawatan, keamanan, dan tenaga kerja untuk gedung di seluruh Indonesia: satu mitra, satu tim di lokasi, satu laporan.'
 WHERE page = 'home' AND section = 'hero' AND title_id = 'Satu mitra tepercaya.';
UPDATE page_sections SET title_id = 'Ceritakan tentang gedung Anda.',
       body_id = 'Kami akan menjadwalkan survei lokasi dan mengirim proposal terperinci. Ingin bicara dulu? Hubungi kami lewat WhatsApp.'
 WHERE page = 'home' AND section = 'quote' AND title_id = 'Ceritakan kebutuhan bisnis Anda.';
UPDATE page_sections SET title_en = 'Five business lines. One partner on site.', title_id = 'Lima lini layanan. Satu mitra di lokasi.'
 WHERE page = 'services' AND section = 'hero' AND title_en = 'Integrated Services for the Way Your Business Operates';
UPDATE page_sections SET title_en = 'Not sure which services your building needs?', title_id = 'Belum yakin layanan apa yang dibutuhkan gedung Anda?'
 WHERE page = 'services' AND section = 'cta' AND title_en = 'Not Sure What Your Operations Need?';
UPDATE page_sections SET title_en = 'Built around how each kind of building runs.',
       title_id = 'Disesuaikan dengan cara setiap gedung beroperasi.',
       body_en = 'A hospital kitchen and a mining camp canteen have different rules. Our experience across the Lippo ecosystem and beyond lets us staff and train for the site you have.',
       body_id = 'Dapur rumah sakit dan kantin site tambang punya aturan berbeda. Pengalaman kami di ekosistem Lippo dan di luarnya membuat kami bisa menyiapkan dan melatih tim sesuai lokasi Anda.'
 WHERE page = 'industries' AND section = 'hero' AND title_en = 'Solutions for the Needs of Every Industry We Serve.';
UPDATE page_sections SET body_en = 'To be Indonesia’s most trusted integrated facility solutions partner.',
       body_id = 'Menjadi mitra solusi fasilitas terpadu paling tepercaya di Indonesia.', is_placeholder = true
 WHERE page = 'about' AND section = 'vision' AND body_en = 'To be a trusted integrated solutions partner that enables businesses to focus on their core operations.';
UPDATE page_sections SET is_placeholder = true
 WHERE page = 'about' AND section = 'vision_mission' AND NOT is_placeholder;
UPDATE page_sections SET body_en = substr(body_en, length('We enable our clients to focus on their core business by taking on the operational services they rely on every day.') + 3),
       body_id = substr(body_id, length('Kami membantu klien fokus pada bisnis inti mereka dengan mengambil alih layanan operasional yang mereka andalkan setiap hari.') + 3)
 WHERE page = 'about' AND section = 'story'
   AND body_en LIKE 'We enable our clients to focus on their core business by taking on the operational services they rely on every day.' || E'\n\n' || '%'
   AND body_id LIKE 'Kami membantu klien fokus pada bisnis inti mereka dengan mengambil alih layanan operasional yang mereka andalkan setiap hari.' || E'\n\n' || '%';
UPDATE page_sections SET title_en = 'Your operations. Our people.', title_id = 'Operasional Anda. Orang-orang kami.'
 WHERE page = 'careers' AND section = 'hero' AND title_en = 'Join the teams that keep essential operations running.';
UPDATE page_sections SET title_en = 'Discuss Your Requirements', title_id = 'Minta penawaran.',
       body_en = 'Tell us about your building. We will arrange a site survey and send an itemised proposal, usually within one working day.',
       body_id = 'Ceritakan tentang gedung Anda. Kami akan menjadwalkan survei lokasi dan mengirim proposal terperinci, biasanya dalam satu hari kerja.'
 WHERE page = 'contact' AND section = 'hero' AND title_en = 'Tell Us What Your Business Needs.';

UPDATE content_items SET title_en = 'Digital reporting & proof of work', title_id = 'Laporan digital & bukti kerja'
 WHERE list_key = 'home_why' AND title_en = 'Transparent Operations, Documented Delivery';
UPDATE content_items SET title_en = '24/7 response', title_id = 'Respons 24/7'
 WHERE list_key = 'home_why' AND title_en = 'A Dedicated Point of Contact';
UPDATE content_items SET body_en = 'Deliver safe, consistent services through trained and supervised teams.',
       body_id = 'Menghadirkan layanan yang aman dan konsisten melalui tim yang terlatih dan diawasi.', is_placeholder = true
 WHERE list_key = 'about_mission' AND body_en = 'Enable clients to focus on their core business through integrated service solutions.';
UPDATE content_items SET body_en = 'Make operations transparent, with a digital record of every job.',
       body_id = 'Membuat operasional transparan, dengan catatan digital untuk setiap pekerjaan.', is_placeholder = true
 WHERE list_key = 'about_mission' AND body_en = 'Deliver safe, reliable and consistent services through capable, accountable teams.';
UPDATE content_items SET body_en = 'Simplify our clients’ operations with one partner for food, facilities, workforce and supply.',
       body_id = 'Menyederhanakan operasional klien dengan satu mitra untuk makanan, fasilitas, tenaga kerja, dan pasokan.', is_placeholder = true
 WHERE list_key = 'about_mission' AND body_en = 'Simplify operations through coordinated delivery, clear accountability and transparent reporting.';

UPDATE industries SET body_en = 'Clinical nutrition and hospital food operations, patient meals and infection-aware cleaning for wards, clinics and laboratories.',
       body_id = 'Gizi klinis dan operasional makanan rumah sakit, makanan pasien, dan kebersihan standar pengendalian infeksi untuk bangsal, klinik, dan laboratorium.'
 WHERE key = 'healthcare' AND body_en LIKE 'Supporting healthcare environments.%';
UPDATE industries SET body_en = 'Canteens, pantries and coffee, cleaning and building upkeep for offices and corporate campuses.',
       body_id = 'Kantin, pantry dan kopi, kebersihan, serta perawatan gedung untuk kantor dan kawasan korporat.'
 WHERE key = 'corporate' AND body_en LIKE 'Offices need well-run workplaces%';
UPDATE industries SET body_en = 'Large-scale retail operational support: large-floor cleaning, security and food solutions for malls and stores.',
       body_id = 'Dukungan operasional ritel skala besar: kebersihan area luas, keamanan, dan solusi makanan untuk mal dan toko.'
 WHERE key = 'retail' AND body_en LIKE 'Malls and stores need clean, safe spaces%';
UPDATE industries SET body_en = 'High-volume canteens, supply chain and site services for plants, mines and remote camps. Camp life, city standard.',
       body_id = 'Kantin berkapasitas besar, rantai pasok, dan layanan site untuk pabrik, tambang, dan camp terpencil. Camp life, city standard.'
 WHERE key = 'industrial' AND body_en LIKE 'Plants, mines and remote camps need%';
UPDATE industries SET body_en = 'Quality dining and operational services across hotels, restaurants and food brands, from concept to execution.',
       body_id = 'Layanan makan dan operasional berkualitas untuk hotel, restoran, dan merek makanan, dari konsep hingga eksekusi.'
 WHERE key = 'hospitality' AND body_en LIKE 'Hotels, restaurants and food brands need%';
UPDATE industries SET body_en = 'Meals, cleaning and safety for schools and higher education campuses.',
       body_id = 'Makanan, kebersihan, dan keamanan untuk sekolah dan kampus pendidikan tinggi.'
 WHERE key = 'education' AND body_en LIKE 'Schools and campuses need healthy meals%';
UPDATE industries SET body_en = 'Common-area cleaning, MEP upkeep and security for residential towers.',
       body_id = 'Kebersihan area umum, perawatan MEP, dan keamanan untuk apartemen.'
 WHERE key = 'residential' AND body_en LIKE 'Residential towers need common areas%';

UPDATE service_lines SET title_en = 'Remote', title_id = 'Remote',
       summary_en = 'Supply, production and on-site teams for mining, industrial and remote sites. Camp life, city standard.',
       summary_id = 'Pasokan, produksi, dan tim di lokasi untuk tambang, industri, dan site terpencil. Camp life, city standard.',
       description_en = 'Camp life, city standard. An integrated supply chain that connects procurement, production, logistics and on-site execution for remote operations.',
       description_id = 'Camp life, city standard. Rantai pasok terpadu yang menghubungkan pengadaan, produksi, logistik, dan eksekusi di lokasi untuk operasi terpencil.',
       lead_en = 'Integrated supply chain excellence: procurement, production, logistics and on-site execution for remote operations. Camp life, city standard.',
       lead_id = 'Rantai pasok terpadu: pengadaan, produksi, logistik, dan eksekusi di lokasi untuk operasi terpencil. Camp life, city standard.'
 WHERE key = 'remote' AND title_en = 'Remote-Managed, On-Site Delivered';
