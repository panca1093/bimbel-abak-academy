INSERT INTO school (name, code, npsn, school_types, alamat, status)
SELECT 'SMAN 1 Jakarta', 'sman1jkt', '20101234', '{"SMA"}', 'Jl. Budi Utomo No.7, Ps. Baru, Kec. Sawah Besar, Kota Jakarta Pusat', 'active'
WHERE NOT EXISTS (SELECT 1 FROM school WHERE code = 'sman1jkt');

INSERT INTO school (name, code, npsn, school_types, alamat, status)
SELECT 'SMAN 3 Bandung', 'sman3bdg', '20202345', '{"SMA"}', 'Jl. Belitung No.8, Merdeka, Kec. Sumur Bandung, Kota Bandung', 'active'
WHERE NOT EXISTS (SELECT 1 FROM school WHERE code = 'sman3bdg');

INSERT INTO school (name, code, npsn, school_types, alamat, status)
SELECT 'SMAN 2 Surabaya', 'sman2sby', '20503456', '{"SMA"}', 'Jl. Wijaya Kusuma No.48, Ketabang, Kec. Genteng, Kota Surabaya', 'active'
WHERE NOT EXISTS (SELECT 1 FROM school WHERE code = 'sman2sby');

INSERT INTO school (name, code, npsn, school_types, alamat, status)
SELECT 'SMAN 5 Yogyakarta', 'sman5ygy', '20404567', '{"SMA"}', 'Jl. Nyi Pembayun No.39, Kotagede, Kota Yogyakarta', 'active'
WHERE NOT EXISTS (SELECT 1 FROM school WHERE code = 'sman5ygy');

INSERT INTO school (name, code, npsn, school_types, alamat, status)
SELECT 'MAN 2 Jakarta', 'man2jkt', '20115678', '{"MA"}', 'Jl. Pengadegan Timur No.2, Pancoran, Kota Jakarta Selatan', 'active'
WHERE NOT EXISTS (SELECT 1 FROM school WHERE code = 'man2jkt');

INSERT INTO school (name, code, npsn, school_types, alamat, status)
SELECT 'SMKN 8 Surabaya', 'smkn8sby', '20526789', '{"SMK"}', 'Jl. Kamboja No.18, Ketabang, Kec. Genteng, Kota Surabaya', 'active'
WHERE NOT EXISTS (SELECT 1 FROM school WHERE code = 'smkn8sby');
