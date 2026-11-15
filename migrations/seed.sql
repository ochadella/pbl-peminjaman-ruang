INSERT INTO users (nim_nip, name, email, password_hash, role) VALUES
    ('ADM001', 'Admin Perpustakaan', 'admin@perpus.test',
     '$2b$10$Vlw.y7lNJkDflR82hi2XL.ICHaBqgFiE73Tb4ouIua9AfVoEd57o6', 'admin'),
    ('PTG001', 'Petugas Perpustakaan', 'petugas@perpus.test',
     '$2b$10$Pt1peVERwjLuZrZT8qKP0uQ2FqG/MrqZFgTCysVtD6un8ErrsJYja', 'petugas'),
    ('230000001', 'Syanda Putri', 'mahasiswa@unair.test',
     '$2b$10$KlnPsfmLSwTAOCbEzTNk6uYqw7DrW5yO9Fdhyp2fYZt1bdpMd/aCm', 'mahasiswa')
ON CONFLICT DO NOTHING

INSERT INTO rooms (code, name, location, capacity, facilities) VALUES
    ('PARTANA-1', 'Partana 1', 'Ruang Partana Lantai 3, Perpustakaan UNAIR Kampus B', 8,
     'TV, AC, Stop Kontak, Meja, Kursi, Papan Tulis'),
    ('PARTANA-2', 'Partana 2', 'Ruang Partana Lantai 3, Perpustakaan UNAIR Kampus B', 8,
     'TV, AC, Stop Kontak, Meja, Kursi, Papan Tulis'),
    ('PARTANA-3', 'Partana 3', 'Ruang Partana Lantai 3, Perpustakaan UNAIR Kampus B', 8,
     'TV, AC, Stop Kontak, Meja, Kursi, Papan Tulis'),
    ('PARTANA-4', 'Partana 4', 'Ruang Partana Lantai 3, Perpustakaan UNAIR Kampus B', 8,
     'TV, AC, Stop Kontak, Meja, Kursi, Papan Tulis')
ON CONFLICT DO NOTHING;

INSERT INTO operating_hours (day_of_week, open_time, close_time, is_open) VALUES
    (1, '08:00', '16:00', TRUE),
    (2, '08:00', '16:00', TRUE),
    (3, '08:00', '16:00', TRUE),
    (4, '08:00', '16:00', TRUE),
    (5, '08:00', '16:00', TRUE),
    (6, '08:00', '12:00', TRUE),
    (0, '08:00', '16:00', FALSE)
ON CONFLICT DO NOTHING;