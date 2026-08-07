<div align="center">

<picture>
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png">
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png">
  <img alt="Divider" src="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png" width="100%">
</picture>

<h1>SLAYER L7</h1>
<h3>Advanced Network Stress Testing & Protocol Exploitation</h3>

<p align="center">
  Hệ thống kiểm tra ứng suất mạng đa giao thức, hiệu năng cao, tối ưu hóa Goroutine và vượt WAF.
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version"></a>
  <a href="#"><img src="https://img.shields.io/badge/Methods-75+-FF4136?style=for-the-badge&logo=codeforces&logoColor=white" alt="Methods"></a>
  <a href="#"><img src="https://img.shields.io/badge/Architecture-Goroutine-2EA44F?style=for-the-badge&logo=go&logoColor=white" alt="Architecture"></a>
  <a href="#"><img src="https://img.shields.io/badge/Proxy-SOCKS5%20%7C%20HTTP-1A1A1A?style=for-the-badge&logo=proxyman&logoColor=white" alt="Proxy"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue?style=for-the-badge&logo=readthedocs&logoColor=white" alt="License"></a>
</p>

<p align="center">
  <a href="#disclaimer">Disclaimer</a> &bull;
  <a href="#tong-quan">Tổng quan</a> &bull;
  <a href="#cai-dat">Cài đặt</a> &bull;
  <a href="#cac-method-tan-cong">Methods</a> &bull;
  <a href="#su-dung">Sử dụng</a> &bull;
  <a href="#troubleshooting">Troubleshooting</a>
</p>

<picture>
  <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png">
  <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png">
  <img alt="Divider" src="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png" width="100%">
</picture>

</div>

## Disclaimer

> [!WARNING]
> **Slayer L7** được thiết kế dành cho các quản trị viên hệ thống, nhà nghiên cứu bảo mật để kiểm tra ứng suất (Stress Testing) trên hệ thống của chính họ. Tôi không chịu trách nhiệm cho bất kỳ hành vi lạm dụng hoặc thiệt hại nào gây ra bởi công cụ này. Việc sử dụng tool vào mục đích tấn công máy chủ không được phép là vi phạm pháp luật.

## Tong Quan

**Slayer L7** không chỉ là một công cụ Flood HTTP thông thường. Nó là một nền tảng kiểm tra ứng suất toàn diện, kết hợp giữa nghệ thuật Bypass WAF, khai thác lỗ hổng Giao thức (Protocol Exploits) và Tối ưu hóa Hiệu năng (Go Concurrency).

- **Tốc độ ánh sáng:** Sử dụng Goroutines của Golang, có thể mở hàng nghìn luồng đồng thời với mức tiêu thụ RAM cực thấp.
- **Đa hình (Polymorphic):** Tự động random hóa User-Agent, Payload, Header và Content-Type để đánh lừa WAF/CDN (Cloudflare, Akamai).
- **Giao diện Neofetch:** CLI hiển thị realtime thống kê hệ thống, Status Code và RPS cực kỳ trực quan.
- **Hỗ trợ Proxy Native:** Tự động parse và xoay vòng SOCKS5, SOCKS4, HTTP proxies.
- **Đa giao thức:** Hỗ trợ tấn công từ Layer 7 (HTTP/HTTPS, WebSocket), Layer 4 (TCP/UDP) cho đến Layer 3 (ICMP) và các giao thức game (Minecraft).

## Cai Dat

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

## Cau Hinh CLI Arguments

Tool sử dụng các flag trực tiếp qua CLI, cho phép tùy chỉnh chi tiết tấn công:

| Flag | Kiểu dữ liệu | Mô tả | Bắt buộc |
| --- | --- | --- | --- |
| `-t` | `string` | URL mục tiêu (VD: `http://example.com` hoặc `http://mc.net:25565`) | ✅ |
| `-m` | `string` | Method tấn công (VD: `mixpost`, `mc_bot`) | ❌ (Mặc định: `httpget`) |
| `-w` | `int` | Số luồng chạy đồng thời (VD: `1000`) | ❌ (Mặc định: `2048`) |
| `-d` | `int` | Thời gian chạy tính bằng giây (VD: `60`) | ❌ (Mặc định: `30`) |
| `-p` | `string` | Đường dẫn tới file proxy txt (VD: `proxy.txt`) | ❌ |
| `-r` | `int` | Delay giữa các request trên 1 luồng (ms) | ❌ (Mặc định: `0`) |
| `-v` | `bool` | In chi tiết lỗi ra console | ❌ |

## Cac Method Tan Cong

Slayer L7 hiện đang sở hữu **hơn 75 method** khác nhau, chia thành 7 nhóm chiến thuật chuyên sâu:

<details>
  <summary><b>Volumetric & HTTP/HTTPS L7 (Tấn công tầng ứng dụng)</b></summary>
  
| Method | Mô tả | Tối ưu cho |
| --- | --- | --- |
| `httpget` | HTTP GET Flood kết hợp Cache-Buster ngẫu nhiên. | Website tĩnh, Blog |
| `httppost` | HTTP POST Flood gửi dữ liệu nặng (Base64, JSON). | Form đăng nhập, đăng ký |
| `apiflood` | Đánh sập API endpoint bằng JSON lồng nhau cực sâu. | REST API, Microservices |
| `httpoptions` | CORS Preflight Flood (OPTIONS request). | API Server, Serverless |
| `httpdelete` | Spam request DELETE ép DB truy vấn quyền. | REST API |
| `httpput` | Gửi request PUT mang theo payload nặng. | API ghi dữ liệu |
| `httphead` | Spam HEAD request đánh lừa Nginx cache. | Bypass băng thông |
| `http_json` | Gửi JSON cấu trúc lồng nhau cực nặng. | NodeJS/PHP backend |
| `http_multipart` | Multipart Upload Flood. | Tàn sát ổ cứng server |
| `http_form_bomb` | Gửi 5000 cặp Key-Value form data. | Webserver parser |
| `http_payload` | Gửi 10,000 bytes raw rác. | Buffer Overflow |
</details>

<details>
  <summary><b>Slow & Exhaustion (Tấn công chậm & Cạn kiệt tài nguyên)</b></summary>
  
| Method | Mô tả | Tối ưu cho |
| --- | --- | --- |
| `rudy` | Slow POST (Khai báo 50MB nhưng gửi 1 byte/s). | Apache, IIS |
| `slowloris` | Giữ kết nối TCP bằng cách gửi Header rác từ từ. | Apache, IIS |
| `chunkpost` | Dùng Transfer-Encoding chunked drip. | Bypass Proxy/WAF |
| `http_slow_read` | Gửi GET, đọc Response 1 byte/giây. | Cạn kiệt thread pool |
| `http_dead_conn` | Giữ connection sống 60s không làm gì. | Webserver worker |
| `http_bad_start` | Gửi Header nửa vời giữ luồng treo. | Webserver parser |
</details>

<details>
  <summary><b>Khai thác Giao thức (Protocol Exploits)</b></summary>
  
| Method | Mô tả | Tối ưu cho |
| --- | --- | --- |
| `rapidreset` | HTTP/2 Rapid Reset (CVE-2023-44487). | Mọi server HTTP/2 |
| `h2continuation`| HTTP/2 Continuation Flood (CVE-2024-2730). | HTTP/2 Load Balancer |
| `http_h2_flood` | Mở nhiều stream HTTP/2 gửi POST. | Vượt giới hạn HTTP/1.1 |
| `smuggle_clte` | HTTP Request Smuggling (CL.TE). | Hệ thống nhiều lớp Proxy |
| `smuggle_tete` | HTTP Request Smuggling (TE.TE). | Hệ thống nhiều lớp Proxy |
| `http_conn_smuggle`| Connection Smuggling keep-alive. | Vượt Rate-limit WAF |
</details>

<details>
  <summary><b>Lạm dụng Header, URL & Bypass WAF/CDN</b></summary>
  
| Method | Mô tả | Tối ưu cho |
| --- | --- | --- |
| `mixpost` | Polymorphic POST (Đổi Content-Type liên tục). | WAF/CDN chặt chẽ |
| `cfbypass` | Cloudflare Bypass bằng Header chuẩn trình duyệt. | Cloudflare Free/Pro |
| `cache_poison` | Web Cache Poisoning (Spam file .css rác). | Varnish, Nginx Cache |
| `wsflood` | WebSocket Flood giữ kết nối vĩnh viễn. | Chat app, Web Game |
| `headerflood` | Gửi 100KB Header rác. | Tràn RAM webserver |
| `cookiebomb` | Gửi 64KB Cookie rác. | Tràn session memory |
| `range` | Range Header Attack chồng chéo. | Video streaming, CDN |
| `malformed` | Long URL Attack (8KB URL). | Crash log parser |
| `http_long_hdr` | Gửi 100+ Header dài. | Buffer Overflow WAF |
| `http_invalid_hdr` | Header không hợp lệ. | WAF regex parser |
| `http_empty` | Gửi request không có data. | Cạn kiệt connection |
| `http_invalid_req` | Gửi method không tồn tại. | Crash HTTP parser |
| `http_ghost` | Gửi CRLF rác. | Lỗi parse frame |
| `http_frag` | Gửi request byte-by-byte chậm. | Bypass Anti-DDoS L7 |
| `http_header_split` | CRLF Injection qua query. | Lừa WAF tạo header giả |
| `http_ssrf` | Gửi body yêu cầu server truy cập IP nội bộ. | AWS Metadata, Localhost |
| `http_rapid_connect` | Mở TCP/TLS rồi đóng ngay. | Cạn kiệt port/Conntrack |
| `http_auth` | Spam Basic Auth rác. | Ép server lookup DB |
| `http_ntlm` | Gửi NTLM token rác. | Microsoft IIS, Exchange |
| `http_cache_maxage` | Spam no-cache bypass CDN. | Đánh thẳng Origin |
| `http_event_stream` | Giữ SSE connection. | SSE Server, Queue |
| `http_poll` | Long Polling spam. | Message Queue |
| `http_payload` | Raw Payload 10KB. | Ức chế I/O server |
</details>

<details>
  <summary><b>Vulnerability Probes (Dò quét lỗ hổng & Log Flooding)</b></summary>
  
| Method | Mô tả | Tối ưu cho |
| --- | --- | --- |
| `xss_probe` | Gửi payload XSS qua query. | Đánh sập WAF log |
| `sqli_probe` | Gửi payload SQL Injection. | Database overload |
| `path_traversal` | Gửi path `../../../etc/passwd`. | Lỗi Directory Traversal |
| `redos` | Regex DoS (`aaaaaaaa!`). | Form validation NodeJS/PHP |
| `graphql_batch` | Gửi 500 query lồng nhau. | GraphQL API |
| `zstd_bomb` | Gửi file nén 10MB (giải nén 10GB). | WAF body parser, API upload |
| `pingback` | XML-RPC Pingback Exploit. | WordPress, Drupal |
</details>

<details>
  <summary><b>Minecraft (Game Server Protocol)</b></summary>
  
| Method | Mô tả | Tối ưu cho |
| --- | --- | --- |
| `mc_ping` | Server List Ping Flood. | Bypass AntiBot |
| `mc_bot` | Login Flood / Bot Join. | AuthMe, BungeeCord |
| `mc_bigpacket` | VarInt Overflow. | Vanilla, Netty config |
| `mc_legacy` | Legacy Ping Exploit (`0xFE 0x01`). | Server cũ |
| `mc_nullping` | Gửi Ping nhưng không đọc Response. | Chiếm slot |
| `mc_handshake_flood`| Flood 100 gói Handshake. | Lỗi protocol |
| `mc_hold` | Gửi Ping, nhận Response, ngủ 30-60s. | Chiếm slot player |
| `mc_data` | Gửi 256 byte rác. | Crash packet decoder |
| `mc_bungee` | BungeeCord Spoofing. | Bypass whitelist |
| `mc_varint` | VarInt 5 byte khai báo dài. | Treo Netty thread |
| `mc_ping_var` | Ping host spoofing. | Spam DB log |
| `mc_data_spam` | Spam packet 0x10. | CPU tính toán vật lý |
| `mc_profile_flood` | 50 Login Start 1 connection. | Mojang API lookup |
| `mc_ext_login` | Extended Login Hold. | Cạn kiệt player slot |
| `mc_account_fill` | Login và giữ 120s+. | Kiểm tra AntiBot |
| `mc_spam_pkt` | Spam packet 256 byte. | Buffer Netty |
| `mc_bad_pkt` | 50 byte rác. | Crash log |
| `mc_random_pkt` | Random Packet ID. | Exception parser |
| `mc_slow_read` | Đọc packet 1 byte/s. | Treo thread write |
</details>

<details>
  <summary><b>Layer 4 (TCP/UDP) & Layer 3 (Network)</b></summary>
  
| Method | Mô tả | Tối ưu cho |
| --- | --- | --- |
| `tcp_connect` | TCP Connection Flood (Giữ 5-10s). | VPS giới hạn connection |
| `tcp_slow` | Giữ kết nối TCP gửi 1 byte. | Mail server, bypass SYN |
| `tcp_payload` | Gửi gói tin 1KB rác qua TCP. | Tường lửa stateless |
| `tcp_large_connect` | Gửi 65KB payload qua TCP. | OOM server |
| `tcp_fragmented` | Gửi TCP 10 byte/chunk. | Firewall reassembly |
| `tcp_half_open` | Bắt tay hoàn chỉnh nhưng không gửi data. | Webserver worker |
| `tcp_urg` | TCP Urgent Pointer Flood. | Windows network stack |
| `tcp_oob` | TCP Out-of-Band Data. | Crash Windows server |
| `tcp_fin` | TCP FIN Flood. | Conntrack daemon |
| `tcp_socket_exhaust`| Giữ TCP rỗng qua proxy 60s. | Proxy port exhaust |
| `udp_flood` | Gửi gói tin UDP 1KB. | Game Server, Voice |
| `udp_amp` | UDP Amplification Flood. | Băng thông L4 |
| `udp_dns` | UDP DNS Flood (Query rác). | Open DNS resolver |
| `udp_memcached` | Memcached UDP Amplification. | Memcached暴露 |
| `dns_query` | TCP/UDP DNS Query Flood. | Authoritative DNS |
| `dns_nx` | DNS NXDOMAIN Flood. | Recursive DNS |
| `icmp_flood` | ICMP Echo Flood. | Băng thông L3 |
| `icmp_large` | ICMP Large Packet (1400 bytes). | MTU mismatch, buffer |
| `ack_flood` | TCP ACK Flood. | Firewall state machine |
| `syn_flood` | TCP SYN Flood. | Conntrack table full |
</details>

## Su Dung

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

Đánh Cloudflare Bypass:
```bash
./slayer -t https://foot.wiki/A2KH5C -m cfbypass -w 2000 -d 100 -p proxy.txt
```

## Troubleshooting

| Lỗi | Nguyên nhân & Cách xử lý |
| --- | --- |
| `Err: 377920` | Proxy rác chết. Vui lòng tải proxy mới hoặc bỏ `-p` để đánh DIRECT. |
| `RPS: 1` | Bị nghẽn port VPS (TIME_WAIT). Giảm số luồng `-w` xuống 500-1000. |
| `mc_bot` không chạy | Server có AntiBot. Thử `mc_hold` hoặc `mc_bigpacket`. |
| Bị chặn sau 100 req | Tường lửa target block IP. Bắt buộc dùng proxy SOCKS5. |
| UI bị lỗi dòng | Terminal không hỗ trợ ANSI Escape. Dùng Linux/Mac hoặc Windows Terminal. |

> [!CAUTION]
> Không nên chạy 5000-10000 luồng Raw TCP (`tcp_connect`) trên VPS 1Core/1GB RAM. Nó sẽ làm sập VPS của bạn trước mục tiêu. Khuyến nghị 500-1500 luồng.

## Dong Gop

Mọi đóng góp (Pull Requests) để tối ưu hóa code, thêm method mới hoặc sửa lỗi đều được hoan nghênh. Vui lòng tuân thủ chuẩn code Go.

## License

Dự án này được phân phối dưới giấy phép **MIT License**. Xem file `LICENSE` để biết chi tiết.

<div align="center">
  <picture>
    <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png">
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png">
    <img alt="Divider" src="https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/rainbow.png" width="100%">
  </picture>
  
  <h3>Made with ❤️ by @iw.uyenn._</h3>
  
  <a href="https://discord.gg/fccfwHzms8">
    <img src="https://img.shields.io/badge/Discord-Tham_gia_may_chu-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="Discord">
  </a>
  
  <p><i>Nếu project này hữu ích, hãy ⭐ Star để ủng hộ nhé!</i></p>
</div>
