"""Generate the Postman collection and local environment from the backend routes."""

import json
import re
from pathlib import Path

HERE = Path(__file__).resolve().parent
BASE = "{{base_url}}"
API = BASE + "/api/v1"


def script(lines):
    return {"listen": "test", "script": {"type": "text/javascript", "exec": lines.splitlines()}}


def guard(flag):
    return {"listen": "prerequest", "script": {"type": "text/javascript", "exec": [
        f'if (pm.environment.get("{flag}") !== "true") pm.execution.skipRequest();'
    ]}}


def request(name, method, path, description, body=None, query=None, save=None, flag=None, auth=True, status=None, headers=None):
    url = API + path if path.startswith("/") else BASE + path
    if query:
        # Empty imported Postman variables may remain unresolved as {{name}}.
        # Keep optional filters out of the default URL; add them in Params when needed.
        url += "?" + "&".join(f"{k}={{{{{v}}}}}" for k, v in query.items() if k not in ("search", "status"))
    req = {"method": method, "url": url, "description": description,
           "header": headers or []}
    if not auth:
        req["auth"] = {"type": "noauth"}
    if body is not None:
        req["header"].append({"key": "Content-Type", "value": "application/json"})
        raw = json.dumps(body, ensure_ascii=False, indent=2)
        for numeric_key in ("service_id", "product_id", "plan_id", "add_on_id", "customer_id", "payment_method_id"):
            raw = raw.replace('"{{' + numeric_key + '}}"', '{{' + numeric_key + '}}')
        req["body"] = {"mode": "raw", "raw": raw,
                       "options": {"raw": {"language": "json"}}}
    if status is None:
        status = 201 if method == "POST" and ("Buat" in name or "Signup" in name) else 200
    checks = [
        f'pm.test("HTTP {status}", () => pm.response.to.have.status({status}));',
        'const result = pm.response.json();',
        'pm.test("Respons sukses", () => pm.expect(result.success).to.eql(true));',
    ]
    if save:
        for key, field in save.items():
            checks += [f'if (pm.response.code === {status} && result.data && result.data.{field} !== undefined) '
                       f'pm.environment.set("{key}", result.data.{field});']
    item = {"name": name, "request": req, "response": [], "event": [script("\n".join(checks))]}
    if path == "/health":
        item["event"] = [script('pm.test("HTTP 200", () => pm.response.to.have.status(200));')]
    if flag:
        item["event"].insert(0, guard(flag))
    required = re.findall(r"{{([a-z_]+)}}", path)
    if body is not None:
        required += re.findall(r"{{([a-z_]+_id)}}", req["body"]["raw"])
    required = list(dict.fromkeys(key for key in required if key != "base_url"))
    if required:
        checks = [f'if (!pm.environment.get("{key}")) pm.execution.skipRequest();' for key in required]
        item["event"].insert(0, {"listen": "prerequest", "script": {"type": "text/javascript", "exec": checks}})
    return item


def folder(name, description, items):
    return {"name": name, "description": description, "item": items}


items = [folder("01 - Kesehatan dan autentikasi", "Jalankan login terlebih dahulu. Signup opsional dan membuat tenant baru.", [
    request("Health", "GET", "/health", "Memastikan API merespons di base_url.", auth=False),
    request("Signup bisnis dan admin", "POST", "/signup", "Membuat tenant dan admin baru. Email harus unik; aktifkan run_signup.",
            {"business_name": "PT Nusantara Digital", "name": "Andi Pratama", "email": "{{signup_email}}", "password": "{{signup_password}}"},
            save={"token": "token"}, flag="run_signup", auth=False, status=201),
    request("Login", "POST", "/login", "Login dengan kredensial environment; menyimpan JWT ke token.",
            {"email": "{{login_email}}", "password": "{{login_password}}"}, save={"token": "token"}, auth=False),
    request("Profil saya", "GET", "/me", "Memeriksa JWT dan profil user aktif."),
])]

items.append(folder("02 - Bisnis dan pengguna", "Akses tenant; perubahan bisnis dan staf membutuhkan Admin.", [
    request("Detail bisnis", "GET", "/businesses/me", "Mengambil bisnis dari JWT."),
    request("Ubah bisnis", "PUT", "/businesses/me", "Mengubah nama bisnis; Admin saja.", {"name": "PT Nusantara Digital", "meta": {}}, flag="run_mutations"),
    request("Daftar pengguna", "GET", "/users", "Daftar anggota tenant; Admin saja."),
    request("Buat staf", "POST", "/users", "Membuat Sales atau Finance. Gunakan email unik; Admin saja.",
            {"name": "Sari Pratama", "email": "{{staff_email}}", "password": "{{staff_password}}", "role": "sales"},
            flag="run_mutations", status=201),
]))

catalog = [
    ("Layanan", "services", "service_id", {"name": "Aplikasi Rapat", "description": "Layanan SaaS", "meta": {}},
     {"name": "Aplikasi Rapat Plus", "description": "Layanan SaaS", "meta": {}}),
    ("Produk", "products", "product_id", {"service_id": "{{service_id}}", "name": "Paket Tim", "description": "Untuk tim kecil", "status": "active", "meta": {}},
     {"name": "Paket Tim Plus", "description": "Untuk tim", "status": "active", "meta": {}}),
    ("Paket", "plans", "plan_id", {"product_id": "{{product_id}}", "name": "Bulanan", "price": 100000, "billing_cycle": "monthly", "trial_days": 0, "meta": {}},
     {"name": "Bulanan Plus", "price": 120000, "billing_cycle": "monthly", "trial_days": 0, "meta": {}}),
    ("Add-on", "add-ons", "add_on_id", {"product_id": "{{product_id}}", "name": "Kapasitas ekstra", "price": 20000, "billing_cycle": "monthly", "meta": {}},
     {"name": "Kapasitas ekstra Plus", "price": 25000, "billing_cycle": "monthly", "meta": {}}),
]

for label, segment, key, create, update in catalog:
    path = "/" + segment
    items.append(folder("03 - Katalog / " + label, "Urutan: buat layanan, produk, paket, lalu add-on. Admin atau Sales untuk menulis.", [
        request("Daftar " + label, "GET", path, "Daftar dengan pagination dan pencarian.", query={"page": "page", "limit": "limit", "search": "search"}),
        request("Buat " + label, "POST", path, "Membuat data dan menyimpan ID ke " + key + ".", create, save={key: "id"}, flag="run_mutations", status=201),
        request("Detail " + label, "GET", path + "/{{" + key + "}}", "Mengambil data menurut ID."),
        request("Ubah " + label, "PUT", path + "/{{" + key + "}}", "Mengubah data terpilih.", update, flag="run_mutations"),
        request("Hapus " + label, "DELETE", path + "/{{" + key + "}}", "Menghapus data. Jalankan terpisah setelah dependensi dilepas.", flag="run_irreversible"),
    ]))

items.append(folder("04 - Kupon", "Admin atau Sales dapat mengubah kupon. Kode kupon harus unik.", [
    request("Daftar kupon", "GET", "/coupons", "Daftar kupon dengan pagination.", query={"page": "page", "limit": "limit", "search": "search"}),
    request("Buat kupon", "POST", "/coupons", "Membuat kupon persentase dan menyimpan coupon_id.",
            {"code": "{{coupon_code}}", "discount_type": "percentage", "discount_value": 10, "max_usage": 100}, save={"coupon_id": "id"}, flag="run_mutations", status=201),
    request("Detail kupon", "GET", "/coupons/{{coupon_id}}", "Mengambil kupon menurut ID."),
    request("Ubah kupon", "PUT", "/coupons/{{coupon_id}}", "Mengubah kupon.",
            {"code": "{{coupon_code}}", "discount_type": "percentage", "discount_value": 15, "max_usage": 100}, flag="run_mutations"),
    request("Hapus kupon", "DELETE", "/coupons/{{coupon_id}}", "Menghapus kupon setelah tidak dipakai.", flag="run_irreversible"),
]))

items.append(folder("05 - Pelanggan", "Admin atau Sales dapat mengubah pelanggan.", [
    request("Daftar pelanggan", "GET", "/customers", "Daftar pelanggan dengan pagination.", query={"page": "page", "limit": "limit", "search": "search"}),
    request("Buat pelanggan", "POST", "/customers", "Membuat pelanggan dan menyimpan customer_id.",
            {"name": "Budi Santoso", "email": "{{customer_email}}", "contact": "{{customer_mobile}}", "meta": {}},
            save={"customer_id": "id"}, flag="run_mutations", status=201),
    request("Detail pelanggan", "GET", "/customers/{{customer_id}}", "Mengambil pelanggan menurut ID."),
    request("Ubah pelanggan", "PUT", "/customers/{{customer_id}}", "Mengubah pelanggan.",
            {"name": "Budi Santoso", "email": "{{customer_email}}", "contact": "{{customer_mobile}}", "meta": {}}, flag="run_mutations"),
    request("Hapus pelanggan", "DELETE", "/customers/{{customer_id}}", "Menghapus pelanggan setelah dependensi dilepas.", flag="run_irreversible"),
]))

items.append(folder("06 - Subscription", "Preview dan perubahan subscription membutuhkan Admin atau Sales. Status trial tidak dapat di-pause.", [
    request("Preview harga", "POST", "/subscriptions/preview", "Menghitung harga tanpa membuat subscription.",
            {"plan_id": "{{plan_id}}", "add_ons": []}),
    request("Daftar subscription", "GET", "/subscriptions", "Filter status: trial, active, paused, cancelled.",
            query={"page": "page", "limit": "limit", "search": "search", "status": "subscription_status"}),
    request("Buat subscription", "POST", "/subscriptions", "Membuat subscription dari customer dan plan; menyimpan subscription_id. Plan trial_days=0 memulai status active.",
            {"customer_id": "{{customer_id}}", "plan_id": "{{plan_id}}", "add_ons": [], "meta": {}},
            save={"subscription_id": "id"}, flag="run_mutations", status=201),
    request("Detail subscription", "GET", "/subscriptions/{{subscription_id}}", "Detail termasuk add-on dan kupon."),
    request("Tambah add-on", "POST", "/subscriptions/{{subscription_id}}/add-ons", "Menambah add-on ke subscription.",
            {"add_on_id": "{{add_on_id}}", "quantity": 1}, flag="run_mutations", status=200),
    request("Lepas add-on", "DELETE", "/subscriptions/{{subscription_id}}/add-ons/{{add_on_id}}", "Melepas add-on; mengubah tagihan berikutnya.", flag="run_irreversible"),
    request("Terapkan kupon", "POST", "/subscriptions/{{subscription_id}}/coupons", "Menerapkan kupon berdasarkan kode.",
            {"coupon_code": "{{coupon_code}}"}, flag="run_mutations", status=200),
    request("Lepas kupon", "DELETE", "/subscriptions/{{subscription_id}}/coupons/{{coupon_id}}", "Melepas kupon.", flag="run_irreversible"),
    request("Pause subscription", "POST", "/subscriptions/{{subscription_id}}/pause", "Hanya untuk subscription active.",
            {"reason": "Atas permintaan pelanggan"}, flag="run_irreversible", status=200),
    request("Resume subscription", "POST", "/subscriptions/{{subscription_id}}/resume", "Hanya untuk subscription paused.", flag="run_irreversible", status=200),
    request("Cancel subscription", "POST", "/subscriptions/{{subscription_id}}/cancel", "Pembatalan permanen; jalankan setelah pengujian pembayaran.",
            {"reason": "Atas permintaan pelanggan"}, flag="run_irreversible", status=200),
]))

items.append(folder("07 - Invoice dan dashboard", "Invoice dibuat oleh alur subscription atau worker billing, bukan endpoint create invoice.", [
    request("Ringkasan dashboard", "GET", "/dashboard/summary", "Metrik tenant saat ini."),
    request("Daftar invoice", "GET", "/invoices", "Filter status: pending, paid, failed, cancelled. Isi invoice_id dari hasil daftar.",
            query={"page": "page", "limit": "limit", "search": "search", "status": "invoice_status"}),
    request("Detail invoice", "GET", "/invoices/{{invoice_id}}", "Invoice beserta item dan payment."),
    request("Tandai lunas manual", "POST", "/invoices/{{invoice_id}}/mark-paid", "Mengubah status pembayaran secara manual; Admin atau Finance saja.", flag="run_irreversible", status=200),
]))

items.append(folder("08 - Checkout publik dan Xendit", "Link dibuat oleh Admin/Finance. GET dan POST checkout publik memakai token mentah dari respons create link. Jangan simpan SDK key dalam koleksi.", [
    request("Buat checkout link", "POST", "/invoices/{{invoice_id}}/checkout-links", "Invoice pending diperlukan. Menyimpan checkout_token dan checkout_url.",
            save={"checkout_token": "token", "checkout_url": "checkout_url"}, flag="run_checkout", status=201),
    request("Lihat checkout publik", "GET", "/checkout/{{checkout_token}}", "Data aman untuk customer tanpa JWT; limit per IP dan token.", auth=False),
    request("Mulai sesi kartu", "POST", "/checkout/{{checkout_token}}/session", "Menghasilkan payment_session_id dan components_sdk_key. Di mode mock, SDK key tidak berlaku di Xendit sesungguhnya.",
            {"name": "Budi Santoso", "email": "{{customer_email}}", "mobile": "{{customer_mobile}}"},
            save={"payment_session_id": "payment_session_id"}, flag="run_checkout", auth=False, status=200),
    request("Webhook Payment Session", "POST", "/webhooks/xendit/payment-sessions", "Hanya untuk payload callback Xendit yang valid. Isi webhook_payload dan callback_token dari lingkungan aman; bukan webhook Payments API v3.",
            None, flag="run_webhook", auth=False, headers=[{"key": "x-callback-token", "value": "{{callback_token}}"},
            {"key": "Content-Type", "value": "application/json"}]),
]))
webhook = items[-1]["item"][-1]
webhook["request"]["body"] = {"mode": "raw", "raw": "{{webhook_payload}}", "options": {"raw": {"language": "json"}}}

items.append(folder("09 - Riwayat pembayaran dan autopay", "Admin atau Finance. Autopay perlu metode Xendit tersimpan, persetujuan pelanggan, dan AUTOPAY_ENABLED pada worker.", [
    request("Riwayat pembayaran", "GET", "/payments", "Filter opsional invoice_id, subscription_id, status; hapus parameter kosong sebelum mengirim bila perlu.",
            query={"page": "page", "limit": "limit", "status": "payment_status"}),
    request("Metode pembayaran pelanggan", "GET", "/customers/{{customer_id}}/payment-methods", "ID metode pembayaran yang telah tersimpan; isi payment_method_id dari respons."),
    request("Pengaturan autopay", "GET", "/subscriptions/{{subscription_id}}/autopay", "Melihat persetujuan autopay saat ini."),
    request("Aktifkan autopay", "PUT", "/subscriptions/{{subscription_id}}/autopay", "Hanya setelah persetujuan nyata; method harus milik pelanggan dan subscription active.",
            {"enabled": True, "payment_method_id": "{{payment_method_id}}", "consent_reference": "{{consent_reference}}",
             "consented_at": "{{consented_at}}", "consent_expires_at": "{{consent_expires_at}}", "max_amount": 200000}, flag="run_autopay"),
    request("Nonaktifkan autopay", "PUT", "/subscriptions/{{subscription_id}}/autopay", "Menghentikan debit baru; debit yang sudah dimulai tetap direkonsiliasi.",
            {"enabled": False}, flag="run_irreversible"),
]))

items.append(folder("10 - Tagihan bisnis ke PayLanggan", "Admin bisnis. Satu paket Rp300.000 per bulan; checkout pertama menyimpan metode pembayaran untuk debit otomatis setelah persetujuan.", [
    request("Paket PayLanggan", "GET", "/platform/plan", "Satu paket tetap Rp300.000 per bulan."),
    request("Status langganan PayLanggan", "GET", "/platform/subscription", "Status, tanggal berlaku, persetujuan debit, dan tagihan pending."),
    request("Daftar tagihan PayLanggan", "GET", "/platform/invoices", "Tagihan bisnis ke PayLanggan, terpisah dari invoice pelanggan."),
    request("Mulai checkout PayLanggan", "POST", "/platform/subscription/checkout", "Hanya kirim setelah admin bisnis menyetujui debit otomatis Rp300.000 tiap bulan. Membuat checkout Xendit dan menyimpan platform_checkout_url.",
            {"auto_renew_consent": True}, save={"platform_checkout_url": "checkout_url"}, flag="run_platform_checkout", status=201),
    request("Batalkan langganan PayLanggan", "POST", "/platform/subscription/cancel", "Menghentikan perpanjangan berikutnya; periode yang sudah dibayar tetap tercatat.", flag="run_irreversible"),
]))

collection = {
    "info": {"name": "PAY-LANGGAN Backend API", "description": "Seluruh route HTTP Backend/internal/routes. Alur: login, katalog, pelanggan, subscription, invoice, checkout. Request mutasi hanya berjalan jika flag terkait bernilai true. Token JWT disimpan di environment lokal setelah login.",
             "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"},
    "auth": {"type": "bearer", "bearer": [{"key": "token", "value": "{{token}}", "type": "string"}]},
    "item": items,
}

values = [
    ("base_url", "http://localhost:8080", "URL API tanpa slash akhir"),
    ("login_email", "admin@example.com", "Email akun tenant yang sudah ada"),
    ("login_password", "", "Password login; isi secara lokal"),
    ("token", "", "JWT diisi otomatis oleh login/signup"),
    ("signup_email", "andi.pratama@example.com", "Email unik untuk signup opsional"),
    ("signup_password", "", "Password signup minimal 6 karakter"),
    ("staff_email", "sari.pratama@example.com", "Email unik staf baru"),
    ("staff_password", "", "Password staf minimal 6 karakter"),
    ("customer_email", "budi.santoso@example.com", "Email pelanggan uji"),
    ("customer_mobile", "+6281234567890", "Nomor E.164 untuk checkout Xendit"),
    ("coupon_code", "HEMAT10", "Kode kupon unik"),
    ("page", "1", "Halaman daftar"), ("limit", "20", "Ukuran halaman"),
    ("search", "", "Pencarian opsional"),
    ("subscription_status", "", "Filter opsional subscription"),
    ("invoice_status", "", "Filter opsional invoice"),
    ("payment_status", "", "Filter opsional pembayaran"),
    *[(key, "", "Diisi otomatis saat create atau manual dari respons daftar") for key in
      ("service_id", "product_id", "plan_id", "add_on_id", "coupon_id", "customer_id", "subscription_id")],
    ("invoice_id", "", "Ambil dari daftar invoice; invoice mungkin dibuat worker billing"),
    ("checkout_token", "", "Token mentah diisi dari create checkout link"),
    ("checkout_url", "", "URL checkout diisi dari create checkout link"),
    ("payment_session_id", "", "Diisi dari start session"),
    ("payment_method_id", "", "Ambil dari daftar metode pembayaran pelanggan"),
    ("platform_checkout_url", "", "URL pembayaran bisnis diisi dari checkout PayLanggan"),
    ("consent_reference", "", "Referensi bukti persetujuan pelanggan"),
    ("consented_at", "", "Waktu persetujuan ISO 8601, tidak boleh di masa depan"),
    ("consent_expires_at", "", "Masa berlaku ISO 8601, harus di masa depan"),
    ("callback_token", "", "Token callback Xendit; jangan commit nilai sebenarnya"),
    ("webhook_payload", "", "Payload callback Xendit yang valid"),
    ("run_signup", "false", "Izinkan signup bisnis baru"),
    ("run_mutations", "false", "Izinkan create/update katalog, pelanggan, subscription"),
    ("run_checkout", "false", "Izinkan membuat checkout link dan sesi"),
    ("run_webhook", "false", "Izinkan pengiriman webhook uji"),
    ("run_autopay", "false", "Izinkan aktivasi autopay setelah consent sah"),
    ("run_platform_checkout", "false", "Izinkan checkout tagihan bisnis setelah persetujuan debit otomatis"),
    ("run_irreversible", "false", "Izinkan delete, lifecycle, mark-paid, dan disable autopay"),
]
environment = {"name": "PAY-LANGGAN Local", "values": [
    {"key": k, "value": v, "type": "default", "enabled": True, "description": d} for k, v, d in values
], "_postman_variable_scope": "environment", "_postman_exported_using": "PAY-LANGGAN generator"}

(HERE / "PAY-LANGGAN.postman_collection.json").write_text(json.dumps(collection, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
(HERE / "PAY-LANGGAN.local.postman_environment.json").write_text(json.dumps(environment, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
