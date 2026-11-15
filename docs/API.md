# Dokumentasi Endpoint RuangKu

Base URL: `http://localhost:3000/api/v1`

Endpoint yang butuh login wajib membawa header `Authorization: Bearer <token dari /auth/login>`.

Semua response memakai format `{ "success": true, "message": "pesan", "data": { } }`.

| Kode | Arti |
|---|---|
| 200 / 201 | Berhasil / data berhasil dibuat |
| 400 | Body atau parameter tidak valid |
| 401 | Belum login atau token tidak valid |
| 403 | Role tidak punya akses |
| 404 | Data tidak ditemukan |
| 409 | Data bentrok (kode ruang dobel, sudah check-in, dll.) |
| 422 | Validasi gagal (format jam salah, password kurang dari 8 karakter, dll.) |

## Daftar endpoint

| Method | URL | Akses | Keterangan |
|---|---|---|---|
| GET | `/health` | publik | Cek server hidup |
| POST | `/auth/register` | publik | Daftar akun mahasiswa |
| POST | `/auth/login` | publik | Login, mendapat token |
| GET | `/auth/me` | login | Profil user yang login |
| GET | `/users` | admin | Daftar semua user |
| PATCH | `/users/:id/role` | admin | Ubah role user |
| GET | `/rooms` | login | Daftar ruang (non-admin hanya ruang aktif) |
| GET | `/rooms/:id` | login | Detail ruang |
| GET | `/rooms/:id/availability?date=YYYY-MM-DD` | login | Jam terpakai dan jam kosong |
| POST | `/rooms` | admin | Tambah ruang |
| PATCH | `/rooms/:id` | admin | Ubah ruang (field opsional) |
| GET | `/operating-hours` | publik | Jam operasional 7 hari |
| PUT | `/operating-hours/:day` | admin | Atur jam operasional (0 = Minggu ... 6 = Sabtu) |
| GET | `/violations/me` | login | Riwayat pelanggaran sendiri |
| GET | `/violations/user/:id` | petugas, admin | Riwayat pelanggaran mahasiswa |
| POST | `/violations` | petugas, admin | Catat pelanggaran |
| POST | `/attendance/check-in` | pemilik booking, petugas, admin | Check-in pakai kode barcode |
| POST | `/attendance/:id/check-out` | pemilik booking, petugas, admin | Check-out booking |
| POST | `/attendance/process-no-show` | petugas, admin | Jalankan pengecekan no-show sekarang |

Endpoint booking (buat, lihat, batal) menyusul dari modul booking.

## Contoh request dan response

### Register
~~~bash
curl -X POST localhost:3000/api/v1/auth/register -H "Content-Type: application/json" -d '{"nim_nip":"434241070","name":"Ocha","email":"ocha@test.com","password":"rahasia123"}'
~~~
Response: `201 "registrasi berhasil"` dengan data user ber-role `mahasiswa`.

### Login
~~~bash
curl -X POST localhost:3000/api/v1/auth/login -H "Content-Type: application/json" -d '{"email":"ocha@test.com","password":"rahasia123"}'
~~~
Response: `200 "login berhasil"` berisi `token` dan `user`. Email atau password salah → `401 "email atau password salah"`.

### Ubah role (admin)
~~~bash
curl -X PATCH localhost:3000/api/v1/users/4/role -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"role":"petugas"}'
~~~

### Tambah ruang (admin)
~~~bash
curl -X POST localhost:3000/api/v1/rooms -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"code":"partana-5","name":"Partana 5","location":"Lantai 3","capacity":6,"facilities":"AC"}'
~~~
Response: `201 "ruang berhasil ditambahkan"`, kode otomatis jadi `PARTANA-5`. Kode sudah dipakai → `409 "kode ruang sudah dipakai"`.

### Ubah ruang (admin)
~~~bash
curl -X PATCH localhost:3000/api/v1/rooms/1 -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"capacity":10,"is_active":true}'
~~~

### Cek ketersediaan ruang
~~~bash
curl "localhost:3000/api/v1/rooms/1/availability?date=2026-11-16" -H "Authorization: Bearer $TOKEN"
~~~
Response: `booked` 10:00–12:00, `free` 08:00–10:00 dan 12:00–16:00. Hari libur → `"perpustakaan tutup pada tanggal ini"`.

### Atur jam operasional (admin)
~~~bash
curl -X PUT localhost:3000/api/v1/operating-hours/6 -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"open_time":"08:00","close_time":"13:00","is_open":true}'
~~~
Jam tutup lebih awal dari jam buka → `422 "jam tutup harus lebih besar dari jam buka"`.

### Catat pelanggaran (petugas/admin)
~~~bash
curl -X POST localhost:3000/api/v1/violations -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"user_id":3,"type":"kerusakan","note":"kursi patah"}'
~~~
Response berisi `violation`, `total_violations`, dan `blocked_until`. `type` hanya `no_show`, `telat_checkout`, atau `kerusakan`. Pelanggaran ke-3 mengisi `blocked_until` 7 hari ke depan.

### Check-in
~~~bash
curl -X POST localhost:3000/api/v1/attendance/check-in -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"checkin_code":"TESA1"}'
~~~
| Kondisi | Response |
|---|---|
| Dalam jendela 15 menit | `200 "check-in berhasil"` |
| Terlalu cepat | `422 "check-in baru bisa dilakukan mulai jam 09:45"` |
| Terlambat | `422 "waktu check-in sudah lewat (batas jam 19:44)"` |
| Sudah check-in | `409 "booking ini sudah check-in"` |

### Check-out
~~~bash
curl -X POST localhost:3000/api/v1/attendance/1/check-out -H "Authorization: Bearer $TOKEN"
~~~
Response: `200 "check-out berhasil"` dengan `late: false`. Bila `late: true`, pelanggaran `telat_checkout` tercatat otomatis.

### Proses no-show manual (petugas/admin)
~~~bash
curl -X POST localhost:3000/api/v1/attendance/process-no-show -H "Authorization: Bearer $TOKEN"
~~~
Response: `{"processed": 1}`. Tanpa dipanggil pun, pengecekan ini berjalan otomatis setiap 1 menit.
