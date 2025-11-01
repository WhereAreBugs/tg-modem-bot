# 📶 Modem Telegram Bot

一个通过 Telegram 机器人远程控制蜂窝网络调制解调器（Modem）的 Go 项目。它允许您随时随地查询设备状态、收发短信、控制移动数据，甚至通过底层AT命令与eSIM交互。

本项目最初为 **Fibocom FM350** 模块设计和测试，但其模块化的引擎设计使其具备扩展以支持其他 Modem 的潜力。

---

## ✨ 主要功能

-   **远程状态监控 (`/status`)**
    -   查询调制解调器的连接状态、运营商名称、注册模式（漫游/归属）。
    -   获取实时网络类型（如 4G/LTE, 5G）。
    -   监控信号质量百分比和 S/N（信噪比/SINR）。
    -   查看当前数据连接的在线时长和 IP 地址。

-   **完整的短信管理**
    -   列出模块内所有短信，并为每条短信分配临时ID (`/sms`)。
    -   根据号码和内容发送短信 (`/sendsms`)。
    -   根据临时ID删除指定短信 (`/deletesms`)。
    -   **自动化**: 实时监听新短信，自动推送到管理员并从模块中删除。

-   **核心设备控制**
    -   一键开启或关闭移动数据连接 (`/data`)。
    -   远程切换物理 SIM 卡槽 (`/switchsim`)。

-   **底层硬件交互**
    -   通过 **AT 命令** 与 eSIM 交互，目前支持查询 ICCID (`/esim list`)。

-   **实时事件通知**
    -   当有电话呼入时，立即向管理员发送通知。

-   **机器人基础功能**
    -   为不同用户（管理员/普通用户）显示不同的命令列表和帮助信息 (`/help`)。
    -   方便用户获取其 Telegram Chat ID (`/getid`)。

---

## 🔧 环境要求

-   **硬件**:
    -   一台运行 Linux 的主机（例如服务器、树莓派等）。
    -   一个兼容的蜂窝网络调制解调器（本项目已在 **Fibocom FM350** 上验证）。
    -   一张（或多张）有效的 SIM 卡或 eSIM。

-   **软件**:
    -   Go 语言环境 (版本 >= 1.18)。
    -   `ModemManager` 服务（Linux 系统中用于管理调制解调器的标准服务）。
    -   `picocom` 或其他串口工具（用于可选的 AT 命令调试）。

-   **配置**:
    -   一个 Telegram 机器人及其 `TOKEN`。
    -   您的个人 Telegram 账户的 `Chat ID`，用于接收管理通知和执行管理员命令。

---

## 🚀 快速开始

1.  **克隆项目**
    ```bash
    git clone https://github.com/WhereAreBugs/tg-modem-bot.git
    cd modem-tg-bot
    ```

2.  **整理依赖**
    ```bash
    go mod tidy
    ```

3.  **配置环境变量**
    在终端中设置您的机器人 Token 和管理员 Chat ID。
    > tips: Chat ID可以通过 /getid 命令 获取。
    ```bash
    export TELEGRAM_BOT_TOKEN="在此处粘贴您的机器人Token"
    export ADMIN_CHAT_ID="在此处粘贴您的Chat ID"
    ```

4.  **编译项目**
    ```bash
    go build -o main ./cmd/main.go
    ```

5.  **运行机器人**
    程序需要权限访问 D-Bus 系统总线和串口设备 (`/dev/wwan0at0` 等)。
    ```bash
    sudo ./main
    ```
    如果一切顺利，您将在终端看到 "Modem 引擎初始化成功" 和 "开始监听 Telegram 更新..." 的日志。现在，您可以向您的机器人发送命令了！

---

## 📖 命令列表

#### 公开命令
-   `/getid` - 获取你当前的 Chat ID
-   `/help` - 显示此帮助信息

#### 管理员命令
-   `/status` - 查询调制解调器详细状态
-   `/sms` - 读取所有短信 (带ID)
-   `/sendsms <号码> <内容>` - 发送短信
-   `/deletesms <ID>` - 删除指定ID的短信
-   `/data <on|off>` - 开启或关闭移动数据
-   `/switchsim <slot>` - 切换SIM卡槽 (例如: `/switchsim 1`)
- 还有更多命令待开发...
---

## 🏗️ 项目结构

-   `cmd/` - 程序主入口 (`main.go`)。
-   `engine/` - 核心引擎，负责与底层硬件和服务交互。
    -   `dbus_mbim/` - 基于 D-Bus 和 ModemManager 的标准功能实现。
    -   `at/` - 独立的 AT 命令处理器，用于与串口直接通信，实现 D-Bus 未暴露的功能（如eSIM）。
-   `commands/` - Telegram 命令的处理器，负责解析和响应用户输入。
-   `automation/` - 后台自动化任务，如短信和来电的 D-Bus 信号监听器。

---

## 📜 授权协议

本项目使用 [MIT](https://opensource.org/licenses/MIT) 授权协议。