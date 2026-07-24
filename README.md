<div align="center">

<img src="https://readme-typing-svg.herokuapp.com?font=JetBrains+Mono&weight=700&size=28&pause=1000&color=a78bfa&center=true&vCenter=true&random=false&width=700&lines=SLAYER+L7;Advanced+Network+Stress+Testing;Layer+4+%26+Layer+7+Exploits;Built+with+Go+Concurrency" alt="Slayer L7" />

<p align="center">
  <strong>Hệ thống kiểm tra ứng suất mạng đa giao thức, hiệu năng cao, tối ưu hóa Goroutine và vượt WAF.</strong>
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"></a>
  <a href="#"><img src="https://img.shields.io/badge/Methods-34+-FF4136?style=for-the-badge&logo=codeforces&logoColor=white" alt="Methods"></a>
  <a href="#"><img src="https://img.shields.io/badge/Architecture-Goroutine-2EA44F?style=for-the-badge&logo=go&logoColor=white" alt="Architecture"></a>
  <a href="#"><img src="https://img.shields.io/badge/Proxy-SOCKS5%20%7C%20HTTP-1A1A1A?style=for-the-badge&logo=proxyman&logoColor=white" alt="Proxy"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue?style=for-the-badge&logo=readthedocs&logoColor=white" alt="License"></a>
</p>

<p align="center">
  <a href="#tong-quan">Tổng quan</a> •
  <a href="#thong-ke-du-an">Thống kê</a> •
  <a href="#cai-dat">Cài đặt</a> •
  <a href="#cau-hinh-api">Cấu hình API</a> •
  <a href="#cac-method-tan-cong">Methods</a> •
  <a href="#su-dung">Sử dụng</a> •
  <a href="#troubleshooting">Troubleshooting</a>
</p>

<img src="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png" width="100%" />

</div>

## ![Disclaimer](https://api.iconify.design/carbon:warning-alt.svg?color=FF4136&width=20) Miễn trừ trách nhiệm

> [!WARNING]
> **Slayer L7** được thiết kế dành cho các quản trị viên hệ thống, nhà nghiên cứu bảo mật để kiểm tra ứng suất (Stress Testing) trên hệ thống của chính họ. Tôi không chịu trách nhiệm cho bất kỳ hành vi lạm dụng hoặc thiệt hại nào gây ra bởi công cụ này. Việc sử dụng tool vào mục đích tấn công máy chủ không được phép là vi phạm pháp luật.

## ![Stats](https://api.iconify.design/carbon:analytics.svg?color=00ADD8&width=20) Thống kê dự án

<div align="center">
  
<img src="https://github-readme-stats.vercel.app/api?username=yinlewoaisuru&repo=Slayer-L7&show_icons=true&theme=blue-green&hide_border=true&count_private=true" alt="GitHub Stats" width="48%" />
<img src="https://github-readme-stats.vercel.app/api/top-langs/?username=yinlewoaisuru&repo=Slayer-L7&layout=compact&theme=blue-green&hide_border=true" alt="Top Languages" width="48%" />

<br/>
<br/>

<img src="https://github-readme-streak-stats.herokuapp.com/?user=yinlewoaisuru&repo=Slayer-L7&theme=blue-green&hide_border=true" alt="Streak Stats" width="70%" />

</div>

## ![Overview](https://api.iconify.design/carbon:cloud.svg?color=white&width=20) Tổng quan

**Slayer L7** không chỉ là một công cụ Flood HTTP thông thường. Nó là một nền tảng kiểm tra ứng suất toàn diện, kết hợp giữa nghệ thuật Bypass WAF, khai thác lỗ hổng Giao thức (Protocol Exploits) và Tối ưu hóa Hiệu năng (Go Concurrency).

- **Tốc độ ánh sáng:** Sử dụng Goroutines của Golang, có thể mở hàng nghìn luồng đồng thời với mức tiêu thụ RAM cực thấp.
- **Đa hình (Polymorphic):** Tự động random hóa User-Agent, Payload, Header và Content-Type để đánh lừa WAF/CDN (Cloudflare, Akamai).
- **Giao diện Neofetch:** CLI hiển thị realtime thống kê hệ thống, Status Code và RPS cực kỳ trực quan.
- **Hỗ trợ Proxy Native:** Tự động parse và xoay vòng SOCKS5, SOCKS4, HTTP proxies.

## ![Installation](https://api.iconify.design/carbon:terminal.svg?color=white&width=20) Cài đặt

### Yêu cầu hệ thống
- [Go (Golang)](https://go.dev/dl/) phiên bản 1.21 trở lên.
- Kết nối Internet ổn định.

### Build từ mã nguồn

```bash
# 1. Clone repository
git clone https://github.com/yinlewoaisuru/Slayer-L7.git
cd Slayer-L7

# 2. Khởi tạo module và tải thư viện
go mod init slayer
go mod tidy

# 3. Biên dịch thành file thực thi
go build -o slayer main.go

# 4. Cấp quyền chạy (Linux/Mac)
chmod +x slayer
```

## ![Config](https://api.iconify.design/carbon:settings.svg?color=white&width=20) Cấu hình & API Payload

Tool sẽ tự động tạo `config.json` trong lần chạy đầu tiên. Bạn có thể chỉnh sửa file này để thay đổi cài đặt mặc định hoặc truyền trực tiếp qua CLI.

```json
{
    "mode": "mix",
    "threads": 1000,
    "concurrent_proxies": 50,
    "delay_after_success": 0,
    "timeout": 15,
    "rate_limit": 0,
    "proxy_types": ["socks5", "socks4", "http"]
}
```

<details>
  <summary><b>CLI Arguments (API Endpoints)</b></summary>

  | Flag | Kiểu dữ liệu | Mô tả | Bắt buộc |
  | --- | --- | --- | --- |
  | `-t` | `string` | URL mục tiêu (VD: `http://example.com` hoặc `http://mc.net:25565`) | ✅ |
  | `-m` | `string` | Method tấn công (VD: `mixpost`, `mc_bot`) | ❌ (Mặc định: `httpget`) |
  | `-w` | `int` | Số luồng chạy đồng thời (VD: `1000`) | ❌ (Mặc định: `2048`) |
  | `-d` | `int` | Thời gian chạy tính bằng giây (VD: `60`) | ❌ (Mặc định: `30`) |
  | `-p` | `string` | Đường dẫn tới file proxy txt (VD: `proxy.txt`) | ❌ |
  | `-r` | `int` | Delay giữa các request trên 1 luồng (ms) | ❌ (Mặc định: `0`) |
  | `-v` | `bool` | In chi tiết lỗi ra console | ❌ |

</details>

## ![Methods](https://api.iconify.design/carbon:code.svg?color=green&width=20) Các Method Tấn Công

Slayer L7 hiện đang sở hữu **34 method** khác nhau, chia thành 8 nhóm chiến thuật:

<details>
  <summary><b>🌐 Volumetric (HTTP Flood)</b></summary>
  
| Method | Mô tả |
| --- | --- |
| `httpget` | HTTP GET Flood kết hợp Cache-Buster ngẫu nhiên. |
| `httppost` | HTTP POST Flood gửi dữ liệu nặng (Base64, JSON). |
| `apiflood` | Đánh sập API endpoint bằng JSON lồng nhau cực sâu. |
</details>

<details>
  <summary><b>⏳ Slow & Exhaustion (Tấn công chậm)</b></summary>
  
| Method | Mô tả |
| --- | --- |
| `rudy` | Slow POST (Khai báo 50MB nhưng gửi 1 byte/s). |
| `slowloris` | Giữ kết nối TCP bằng cách gửi Header rác từ từ. |
| `chunkpost` | Dùng Transfer-Encoding chunked drip. |
| `tcp_slow` | Giữ kết nối TCP sống 30-60s gửi 1 byte. |
</details>

<details>
  <summary><b>🚀 Khai thác Giao thức (Protocol Exploits)</b></summary>
  
| Method | Mô tả |
| --- | --- |
| `rapidreset` | HTTP/2 Rapid Reset (CVE-2023-44487). |
| `h2continuation`| HTTP/2 Continuation Flood (CVE-2024-2730). |
| `smuggle_clte` | HTTP Request Smuggling (CL.TE). |
</details>

<details>
  <summary><b>🛡️ Lạm dụng Header & URL</b></summary>
  
| Method | Mô tả |
| --- | --- |
| `headerflood` | Gửi 100KB Header rác. |
| `cookiebomb` | Gửi 64KB Cookie rác. |
| `range` | Range Header Attack chồng chéo. |
| `malformed` | Long URL Attack (8KB URL). |
</details>

<details>
  <summary><b>💣 Vượt WAF/CDN (Bypass)</b></summary>
  
| Method | Mô tả |
| --- | --- |
| `mixpost` | Polymorphic POST (Đổi Content-Type liên tục). |
| `cfbypass` | Cloudflare Bypass bằng Header chuẩn trình duyệt. |
| `cache_poison` | Web Cache Poisoning. |
| `wsflood` | WebSocket Flood giữ kết nối vĩnh viễn. |
</details>

<details>
  <summary><b>🎮 Minecraft (Game Server)</b></summary>
  
| Method | Mô tả |
| --- | --- |
| `mc_ping` | Server List Ping Flood (Giữ kết nối 10-15s). |
| `mc_bot` | Login Flood / Bot Join (Giữ kết nối 15-30s). |
| `mc_bigpacket` | Khai báo length lớn, gửi data nhỏ làm treo Netty. |
| `mc_legacy` | Legacy Ping Exploit (`0xFE 0x01`). |
| `mc_nullping` | Gửi gói Ping nhưng không đọc Response. |
| `mc_handshake_flood`| Flood 100 gói Handshake trong 1 kết nối. |
| `mc_hold` | Gửi Ping, nhận Response, ngủ 30-60s. |
| `mc_data` | Gửi data rác 256 byte crash packet decoder. |
</details>

<details>
  <summary><b>🔌 Layer 4 (TCP/UDP)</b></summary>
  
| Method | Mô tả |
| --- | --- |
| `tcp_connect` | TCP Connection Flood (Giữ 5-10s). |
| `tcp_payload` | Gửi gói tin 1KB rác qua TCP. |
| `udp_flood` | Gửi gói tin UDP 1KB không kết nối. |
</details>

## ![Usage](https://api.iconify.design/carbon:play-outline.svg?color=white&width=20) Sử dụng

### Cú pháp lệnh chuẩn

```bash
./slayer -t <URL_MỤC_TIÊU> -m <METHOD> -w <SỐ_LƯỜNG> -d <THỜI_GIAN_GIÂY> [-p <FILE_PROXY>]
```

### Ví dụ thực tế

Đánh HTTP GET Flood trực tiếp:
```bash
./slayer -t https://example.com -m httpget -w 2000 -d 60
```

Đánh Minecraft Bot Join bằng Proxy:
```bash
./slayer -t http://mc.example.com:25565 -m mc_bot -w 1000 -d 120 -p proxy.txt
```

Đánh Layer 4 UDP Flood:
```bash
./slayer -t http://123.45.67.89:8080 -m udp_flood -w 3000 -d 60
```

## ![Troubleshooting](https://api.iconify.design/carbon:bug.svg?color=red&width=20) Troubleshooting

| Lỗi | Nguyên nhân & Cách xử lý |
| --- | --- |
| `Err: 377920` | Proxy rác chết. Vui lòng tải proxy mới hoặc bỏ `-p` để đánh DIRECT. |
| `RPS: 1` | Bị nghẽn port VPS (TIME_WAIT). Giảm số luồng `-w` xuống 500-1000. |
| `mc_bot` không chạy | Server có AntiBot. Thử `mc_hold` hoặc `mc_bigpacket`. |
| Bị chặn sau 100 req | Tường lửa target block IP. Bắt buộc dùng proxy SOCKS5. |
| UI bị lỗi dòng | Terminal không hỗ trợ ANSI Escape. Dùng Linux/Mac hoặc Windows Terminal. |

> [!CAUTION]
> Không nên chạy 5000-10000 luồng Raw TCP (`tcp_connect`) trên VPS 1Core/1GB RAM. Nó sẽ làm sập VPS của bạn trước mục tiêu. Khuyến nghị 500-1500 luồng.

## ![Contribution](https://api.iconify.design/carbon:user-favorite.svg?color=white&width=20) Đóng góp

Mọi đóng góp (Pull Requests) để tối ưu hóa code, thêm method mới hoặc sửa lỗi đều được hoan nghênh. Vui lòng tuân thủ chuẩn code Go.

## ![License](https://api.iconify.design/carbon:document.svg?color=white&width=20) License

Dự án này được phân phối dưới giấy phép **MIT License**. Xem file `LICENSE` để biết chi tiết.

<div align="center">
  <img src="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png" width="100%" />
  
  <h3>Made with ❤️ by @iw.uyenn._</h3>
  
  <a href="https://discord.gg/fccfwHzms8">
    <img src="https://img.shields.io/badge/Discord-Tham_gia_may_chu-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="Discord">
  </a>
  
  <p><i>Nếu project này hữu ích, hãy ⭐ Star để ủng hộ nhé!</i></p>
</div>
