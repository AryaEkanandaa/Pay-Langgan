# Demo drop-in statis

Isi konfigurasi langsung di `index.html`: alamat skrip frontend, ID integrasi bisnis, alamat API backend, dan ID plan publik. Tidak perlu mengisi variabel environment, nama website, URL website, atau menjalankan `server.mjs`. ID integrasi bersifat publik dan dapat digunakan dari halaman mana pun; jangan perlakukan ID ini sebagai secret key.

## Langkah singkat

1. Di dashboard bisnis, buka **Pengaturan → Integrasi Website**. Pastikan ada plan aktif dan buat ID integrasi jika belum ada.
2. Buka `index.html`, lalu ganti empat nilai `GANTI-DENGAN-...`:
   - `src`: URL publik frontend PayLanggan yang aktif, ditambah `/paylanggan.js`.
   - `data-pl-site`: ID integrasi `pl_site_...` dari dashboard.
   - `data-pl-api-url`: URL publik backend PayLanggan yang aktif.
   - `data-pl-plan`: ID plan publik `pl_plan_...` yang disalin dari dashboard.
3. Dari folder utama proyek, jalankan server file statis bawaan Python:

   ```powershell
   python -m http.server 4177 --directory .\examples\dropin-simple
   ```

4. Buka `http://localhost:4177` dan klik tombol langganan. Port 4176 digunakan oleh frontend PayLanggan.

Server file statis tetap diperlukan agar browser membuka halaman melalui HTTP. Ini bukan `server.mjs` aplikasi dan tidak memerlukan backend tambahan. Jangan membuka `index.html` langsung dengan `file://`.

Untuk mencoba pembayaran Xendit Test Mode, URL frontend dan backend pada HTML harus berupa tunnel HTTPS yang masih aktif. ID integrasi publik tidak dibatasi pada nama atau URL website tertentu.
