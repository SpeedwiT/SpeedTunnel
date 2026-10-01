# ⚡ Speed Tunnel

   ## تانل قدرتمند، پایدار و ضد اختلال برای سرورهای اوبونتو — با ترنسپورت‌های اختصاصی

[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/SpeedwiT/SpeedTunnel)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://golang.org)
[![Ubuntu](https://img.shields.io/badge/Ubuntu-20.04%20%7C%2022.04%20%7C%2024.04-E95420.svg)](https://ubuntu.com)
[![Channel](https://img.shields.io/badge/Telegram-@Speedw__IT-26A5E4.svg)](https://t.me/Speedw_IT)
[![Support](https://img.shields.io/badge/Support-@SpeedwIT-26A5E4.svg)](https://t.me/SpeedwIT)

**Github:** https://github.com/SpeedwiT/SpeedTunnel  
**Channel:** [@Speedw_IT](https://t.me/Speedw_IT) — **Support:** [@SpeedwIT](https://t.me/SpeedwIT)

---

## ✨ ویژگی‌ها

- 🚀 **۴ ترنسپورت اختصاصی** — فراتر از WebSocket/gRPC، طراحی شده برای شبکهٔ ایران
- 🔒 **اسپوف کامل** — SNI Spoof + Host Spoof + Packet Fragmentation + Padding
- 🔄 **پایدار در اختلال شدید** — KeepAlive + Auto Reconnect + Reverse Tunnel برای قطعی اینترنت بین‌الملل
- 🌐 **بدون وابستگی به UDP** — همه چیز روی TCP (چون UDP در ایران ضعیف/مسدود است)
- 📦 **چند تانل همزمان** — هر تانل پورت و ترنسپورت و SNI مجزا
- 🎨 **منوی ترمینالی حرفه‌ای رنگی** — مدیریت کامل از ترمینال
- 🤖 **ربات تلگرام** — نمایش وضعیت لحظه‌ای تانل‌ها
- ⚙️ **سرویس systemd** — اجرای خودکار بعد از ریبوت
- 🔄 **آپدیت از گیتهاب** — یک کلیک از داخل منو

## 🧬 ترنسپورت‌ها

| ترنسپورت | پروتکل پایه | کاربرد |
|---|---|---|
| **SpeedTLS** ⭐ | TCP + Fake TLS 1.3 + Fragmentation | **پیشنهادی** — خودش را جای ترافیک Cloudflare جا می‌زند، پرسرعت و ضد DPI |
| **SpeedHTTP** | TCP + Fake HTTP/2 Upgrade | شبیه ترافیک عادی مرورگر، فوق‌العاده پایدار |
| **SpeedReverse** | Reverse TCP (ایران → خارج) | **وقتی سرور ایران اینترنت بین‌الملل ندارد** — ایران خودش به خارج وصل می‌شود |
| **SpeedICMP** | فالبک اضطراری | برای شرایط اختلال شدید ملی |

> همهٔ ترنسپورت‌ها دارای **Obfuscation**, **KeepAlive (15s)** و **Auto Reconnect با Backoff** هستند.

---

## 🚀 نصب سریع (Ubuntu 20/22/24)

### یک‌خطی (پیشنهادی)

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/SpeedwiT/SpeedTunnel/main/install.sh)
```

بعد از نصب وارد منو شوید:

```bash
sudo speedtunnel-menu
# یا
sudo st
```

### نصب دستی از سورس

```bash
git clone https://github.com/SpeedwiT/SpeedTunnel.git
cd SpeedTunnel
go build -o speedtunnel ./cmd/speedtunnel
sudo cp speedtunnel /usr/local/bin/
sudo cp scripts/menu.sh /usr/local/bin/speedtunnel-menu
sudo chmod +x /usr/local/bin/speedtunnel-menu
sudo ln -sf /usr/local/bin/speedtunnel-menu /usr/local/bin/st
sudo mkdir -p /etc/speedtunnel
echo '{"version":"1.0.0","tunnels":[]}' | sudo tee /etc/speedtunnel/config.json
sudo cp systemd/speedtunnel.service /etc/systemd/system/speedtunnel.service
sudo systemctl daemon-reload
sudo systemctl enable --now speedtunnel
```

---

## 🎮 منوی ترمینالی

```
╔════════════════════════════════════════════════════════════╗
║  ░░█▀▀ █▀█ █▀▀ █▀▀ █▀▄░░▀█▀ █░█ █▀█ █▀█ █▀▀ █░░            ║
║  ░░▀▀█ █▀▀ █▀▀ █▀▀ █░█░░░█░ █░█ █░█ █░█ █▀▀ █░░            ║
║  ░░▀▀▀ ▀░░ ▀▀▀ ▀▀▀ ▀▀░░░░▀░ ▀▀▀ ▀░▀ ▀░▀ ▀▀▀ ▀▀▀            ║
║   ⚡ Speed Tunnel v1.0.0  — Fast • Secure • Anti-DPI      ║
║   Github: https://github.com/SpeedwiT/SpeedTunnel        ║
║   Channel: @Speedw_IT  Support: @SpeedwIT                ║
╚════════════════════════════════════════════════════════════╝

  منوی اصلی  ● فعال (running)  Config: /etc/speedtunnel/config.json

  1) ایجاد تانل جدید          — Create Tunnel
  2) مدیریت تانل‌ها            — List / Edit / Delete / Toggle
  3) مشاهده وضعیت              — Status & Health Check
  4) تنظیمات ربات تلگرام        — Telegram Bot
  5) نمایش لاگ سرویس            — Logs
  6) آپدیت از گیتهاب             — Update from Github
  7) نمایش کانفیگ خام            — Show Config
  8) ریستارت سرویس
  9) حذف Speed Tunnel
  0) خروج
```

### مدیریت تانل

- **ایجاد:** نام، نقش (iran/kharej)، ترنسپورت، ListenPort، RemotePort، آدرس سرور خارج، ControlPort، Secret، SNI
- **ویرایش:** تغییر SNI و پورت در لحظه
- **حذف / فعال-غیرفعال:** هر تانل جداگانه
- **وضعیت:** نمایش سرویس، تعداد تانل‌ها، TCP Ping هر پورت، لاگ systemd

---

## ⚙️ نحوهٔ کار (معماری)

```
[کاربر] → [سرور خارج: ListenPort] ──(Multiplexer/Forward)──▶ [تونل کنترل: ControlPort] ──▶ [سرور ایران] → [سرویس لوکال: RemotePort]
                              ▲                                                        │
                              │  SpeedTLS / SpeedHTTP / SpeedReverse (TCP + Spoof)     │
                              └────────────────────────────────────────────────────────┘
```

- سرور **خارج (kharej)** روی `ControlPort` به تانل گوش می‌دهد و روی `ListenPort` به کاربران سرویس می‌دهد.
- سرور **ایران (iran)** به `ControlPort` خارج وصل می‌شود (حتی با Reverse).
- هر اتصال کاربر به صورت یک **Stream مالتی‌پلکس** از داخل یک TCP کنترل عبور می‌کند (۱۰+ اتصال همزمان روی یک TCP).
- ترافیک کنترل با **Fake TLS Hello + SNI** و **Fragmentation** از DPI عبور می‌کند.

### حالت قطعی اینترنت بین‌الملل

از **SpeedReverse** استفاده کنید: ایران خودش به خارج وصل می‌شود (حتی از طریق IPهای سفید/ملی). کافی است هر دو طرف `transport: speedreverse` داشته باشند.

---

## 📝 کانفیگ

فایل: `/etc/speedtunnel/config.json`

```json
{
  "version": "1.0.0",
  "bot_token": "",
  "bot_admin_id": 0,
  "tunnels": [
    {
      "id": "st-abc123",
      "name": "my-tunnel",
      "role": "kharej",
      "transport": "speedtls",
      "listen_port": 443,
      "remote_port": 443,
      "remote_addr": "0.0.0.0",
      "control_port": 7000,
      "secret": "your-secret-key",
      "sni": "www.digikala.com",
      "enabled": true,
      "created_at": "2026-10-01T00:00:00Z"
    },
    {
      "id": "st-abc123",
      "name": "my-tunnel",
      "role": "iran",
      "transport": "speedtls",
      "listen_port": 443,
      "remote_port": 443,
      "remote_addr": "YOUR_KHAREJ_IP",
      "control_port": 7000,
      "secret": "your-secret-key",
      "sni": "www.digikala.com",
      "enabled": true,
      "created_at": "2026-10-01T00:00:00Z"
    }
  ]
}
```

> **نکته:** مقدار `secret` و `control_port` و `sni` باید در هر دو سرور یکسان باشد. `listen_port` پورتی است که کاربران به آن وصل می‌شوند (روی خارج)، و `remote_port` پورتی است که سرویس اصلی روی ایران روی آن در حال اجراست (مثلا پنل X-UI).

متغیر محیطی برای تست:

```bash
SPEEDTUNNEL_CONFIG=/tmp/my.json speedtunnel run
```

---

## 🤖 ربات تلگرام

۱. از [@BotFather](https://t.me/BotFather) توکن بگیرید.  
۲. در منو گزینه **۴) تنظیمات ربات تلگرام** را بزنید و `bot_token` و `bot_admin_id` (آیدی عددی خودتان) را وارد کنید.  
۳. سرویس را ریستارت کنید (گزینه ۸).

دستورات ربات:

- `/status` — وضعیت همهٔ تانل‌ها + Ping + SNI + پورت
- `/help` — راهنما

> ربات فقط به `bot_admin_id` پاسخ می‌دهد. اگر `0` باشد، به همه پاسخ می‌دهد (فقط برای تست).

---

## 🛡️ اسپوف و ضد DPI

- **SNI Spoof:** مقدار `sni` را روی دامنهٔ ایرانی پرترافیک بگذارید: `www.digikala.com`, `www.cafebazaar.ir`, `arvancloud.ir`
- **Fragmentation:** هر Write به چانک‌های ۴۰۰–۱۲۰۰ بایتی با تاخیر ۰–۵ms خرد می‌شود
- **KeepAlive:** هر ۱۵ ثانیه Ping/Pong برای جلوگیری از Timeout شدن توسط فایروال
- **Padding:** هدر احراز هویت با طول تصادفی

---

## 📋 دستورات سرویس

```bash
sudo systemctl status speedtunnel
sudo systemctl restart speedtunnel
sudo systemctl stop speedtunnel
sudo journalctl -u speedtunnel -f

speedtunnel version
speedtunnel run        # اجرای هر دو نقش بر اساس config
speedtunnel server     # فقط kharej
speedtunnel client     # فقط iran
speedtunnel bot        # فقط ربات
```

---

## 🔄 آپدیت

از داخل منو گزینه **۶) آپدیت از گیتهاب** را بزنید، یا:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/SpeedwiT/SpeedTunnel/main/install.sh) --update
# یا دستی
cd SpeedTunnel && git pull && sudo bash install.sh --update
```

---

## 🧪 تست

```bash
go vet ./...
go build -o /tmp/speedtunnel ./cmd/speedtunnel
# تست E2E: یک Echo روی 19900 و تانل روی 29900
SPEEDTUNNEL_CONFIG=/tmp/test.json /tmp/speedtunnel run
```

تست‌های انجام‌شده:

- ✅ `go vet` پاس
- ✅ بیلد موفق (7.4M باینری استاتیک)
- ✅ E2E SpeedTLS — Echo از طریق Forward
- ✅ E2E SpeedHTTP — Echo از طریق Forward
- ✅ Multi-connection (3 اتصال همزمان روی یک کنترل)

---

## 📁 ساختار پروژه

```
SpeedTunnel/
├── cmd/speedtunnel/      # نقطهٔ ورود باینری
├── internal/
│   ├── config/           # مدیریت config.json
│   ├── crypto/           # کلید و رمزنگاری
│   ├── tunnel/           # Multiplexer + Protocol + Server/Client
│   ├── transport/        # SpeedTLS / SpeedHTTP / Reverse + Fragmentation + Spoof
│   ├── forward/          # Port Forwarding
│   ├── bot/              # ربات تلگرام
│   └── utils/            # ابزار عمومی
├── scripts/menu.sh       # منوی ترمینالی حرفه‌ای
├── systemd/speedtunnel.service
├── install.sh            # نصب یک‌خطی
└── README.md
```

---

## 📄 مجوز

MIT — ساخته شده با ❤️ توسط [@Speedw_IT](https://t.me/Speedw_IT)

---

## 🙏 پشتیبانی

- کانال: [@Speedw_IT](https://t.me/Speedw_IT)
- پشتیبانی: [@SpeedwIT](https://t.me/SpeedwIT)
- گیتهاب: [github.com/SpeedwiT/SpeedTunnel](https://github.com/SpeedwiT/SpeedTunnel)

اگر پروژه به کارتان آمد ⭐ بدهید!
