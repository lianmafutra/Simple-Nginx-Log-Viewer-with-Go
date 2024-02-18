Simple Nginx Log Viewer From Local Path File
======================================

[![Go Reference](https://pkg.go.dev/badge/github.com/your-username/your-app.svg)](https://pkg.go.dev/github.com/your-username/your-app)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

### Features
- Web Dashboard
- Filter By Range Date
- Total Request
- RPS (Request Per Second)
- RPM (Request Per Minute)
- Request URI/URL
- User Agent, HTTP Response Code , more ...

### Planned Features
- Parse to CSV

### Requirements 

1. Default Regex = ``^(\S+) - \[([^\]]+)\] "(\S+) (\S+) (\S+)" (\d+) (\d+) - "([^"]+)" - (\d+\.\d+)$``

2. Nginx Log  = `119.235.212.226 - [18/Feb/2024:21:52:16 +0700] "GET / HTTP/2.0" 200 1644 - "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0" - 0.058`

<b> You can custom for your specific nginx log format <b>


### Usage
1. Put `nginx.log `file in same directory main.go
2. You Must edit Specific Range Date & Path File Nginx Log in `main.go` file
3. Running file `main.go` with command `go run main.go`
4. Open Web `http://localhost:8080`

### Screenshoot
![log](https://github.com/lianmafutra/Simple-Nginx-Log-Viewer-with-Go/assets/15800599/c6e8f244-43ae-4004-ae30-9da8dcb55382)

![New Project (1)](https://github.com/lianmafutra/Simple-Nginx-Log-Viewer-with-Go/assets/15800599/cb5e9b42-fc62-451a-b073-5025c1ba5294)

![New Project](https://github.com/lianmafutra/Simple-Nginx-Log-Viewer-with-Go/assets/15800599/35c9c9b1-c9a7-4f2e-9878-64f5a92e66d0)
