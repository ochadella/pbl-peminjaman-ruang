# Rancangan Sistem RuangKu

RuangKu adalah backend sistem peminjaman ruang diskusi perpustakaan (Ruang Partana). Mahasiswa bisa booking ruang, check-in dengan scan barcode, lalu check-out. Petugas dan admin mengelola ruang, jam operasional, dan pelanggaran.

## Aktor dan hak akses

| Fitur | Mahasiswa | Petugas | Admin |
|---|:---:|:---:|:---:|
| Register, login, lihat profil | ✓ | ✓ | ✓ |
| Lihat ruang dan ketersediaannya | ✓ | ✓ | ✓ |
| Booking dan batal booking | ✓ | | |
| Check-in (scan barcode) dan check-out | ✓ (booking sendiri) | ✓ | ✓ |
| Lihat riwayat pelanggaran sendiri | ✓ | ✓ | ✓ |
| Catat pelanggaran, lihat pelanggaran mahasiswa | | ✓ | ✓ |
| Jalankan pengecekan no-show manual | | ✓ | ✓ |
| Tambah dan ubah ruang | | | ✓ |
| Atur jam operasional | | | ✓ |
| Lihat user dan ubah role | | | ✓ |

## Alur status booking

~~~mermaid
stateDiagram-v2
    [*] --> booked: mahasiswa booking
    booked --> checked_in: scan barcode (15 menit sebelum s/d sesudah mulai)
    booked --> cancelled: mahasiswa batal
    booked --> no_show: lewat 15 menit tanpa check-in
    checked_in --> completed: check-out
    no_show --> [*]: tercatat pelanggaran no_show
    completed --> [*]: telat lebih dari 15 menit = pelanggaran telat_checkout
    cancelled --> [*]
~~~

## Aturan bisnis

1. **Jendela check-in:** check-in hanya bisa 15 menit sebelum sampai 15 menit sesudah jam mulai.
2. **No-show otomatis:** booking yang belum check-in sampai 15 menit setelah jam mulai diubah menjadi `no_show` oleh worker yang berjalan setiap 1 menit, lalu tercatat pelanggaran `no_show`.
3. **Telat check-out:** check-out lebih dari 15 menit setelah jam selesai otomatis tercatat pelanggaran `telat_checkout`.
4. **Blokir:** 3 pelanggaran membuat mahasiswa diblokir booking selama 7 hari (`users.blocked_until`).
5. **Batas booking:** maksimal 1 booking aktif per hari dan durasi maksimal 2 jam (diterapkan di modul booking).
6. **Ketersediaan:** jam kosong dihitung dari jam operasional hari itu dikurangi booking berstatus `booked`, `checked_in`, dan `completed`. Booking `cancelled` dan `no_show` tidak menghalangi ruang.
7. Ruang yang dinonaktifkan admin tidak tampil untuk mahasiswa dan petugas.

## Arsitektur (Clean Architecture)

~~~
main.go                 load config, koneksi DB, jalankan server
config/                 baca .env
database/               koneksi PostgreSQL (pgxpool)
route/                  daftar URL dan middleware, tanpa logika
middleware/             cek token JWT dan cek role
app/model/              struct data dan request
app/repository/         query SQL ke database (pgx)
app/service/            logika bisnis + penerima request Fiber
helper/                 format response, error, JWT
migrations/             SQL pembuatan tabel dan seed data
~~~

Alur request: **route → middleware → service → repository → database**.

Sesuai ketentuan tugas, proyek ini **tidak memakai folder controller atau handler**. Fungsi yang menerima request (`*fiber.Ctx`) langsung ditaruh di `app/service`, dan `route/route.go` hanya mendaftarkannya.

## Teknologi

Go, Fiber v2, PostgreSQL, pgx v5, JWT (golang-jwt v5, HS256, berlaku 24 jam), bcrypt.
