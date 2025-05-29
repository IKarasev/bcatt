
# <img src="https://github.com/IKarasev/bcatt/releases/download/Images/logo320.png" alt="" width="35"/> BCATT - Prototype System for Simulation and Visualization of Attacks on Blockchain Systems

## Description

This solution for simulating and visualizing attacks on blockchain systems includes:
- construction and simulation of a blockchain network with a specified topology and PoW consensus algorithm;
- address-based and random transaction generation;
- algorithms for validating blocks and transactions;
- user web interface;
- control and monitoring of the simulation during execution;
- ability to manually conduct attacks;
- real-time visualization of a network heatmap and the blockchain graph with its branches;

## Dependencies

**Go Packages**

| Package   | Description    |
|--------------- | --------------- |
| [templ](https://github.com/a-h/templ)          | HTML template language for Go with excellent developer tools |
| [gogost](https://github.com/ddulesov/gogost)   | Pure Go library for GOST cryptographic functions |
| [echo](https://github.com/labstack/echo)       | High-performance, extensible, and minimalist web framework for Go |
| [xid](https://github.com/rs/xid)               | Library for generating unique IDs |
| [yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) | Go library for working with YAML files |

**Web Libraries**

| Library   | Description    |
|--------------- | --------------- |
| [HTMX](https://htmx.org/)  | JavaScript library for adding dynamic behavior to web pages |
| [D3.js](https://d3js.org/) | JavaScript library for creating visualizations |
| [Tailwind CSS](https://tailwindcss.com/) | CSS utility framework |
| [daisyUI](https://daisyui.com/) | Plugin for Tailwind CSS |

## Configuration

System parameters can be configured via environment variables or a YAML configuration file. Settings in the YAML file take precedence over environment variables.

### ENV

| Variable   | Default Value    | Description |
| ---- | --- | ---- |
| BCATT_COINBASE_START | 1000000 | Initial balance of the network’s base wallet |
| BCATT_MINE_DIFF | 20 | Mining difficulty parameter |
| BCATT_HTTP_ADDR | 127.0.0.1 | Address for the web server |
| BCATT_HTTP_PORT | 8080 | Web server port |
| BCATT_RSS_UPDATE | 500 | RSS update interval in milliseconds |
| BCATT_OP_PAUSE_MILISEC | 100 | Pause between network actor operations |
| BCATT_WITH_LOG | true | Log output to console |
| BCATT_NODE_NUM | 3 | Number of nodes in the network (each node also has a wallet) |
| BCATT_WALLET_NUM | 1 | Number of additional wallets |

### config.yaml

```yaml
web:
  address: "127.0.0.1"    # IP address for web server
  port: 8080              # Port for web server
  rss_update_time: 100    # Time period for RSS queue update in milliseconds

emulator:
  op_pause: 1             # Pause in milliseconds between operations
  with_log: true          # Print log to console
  start_utxo:             # Settings for generating starting UTXO list
    active: true          # Use starting UTXO generation
    all: true             # Generate UTXOs for all wallets
    wallets:              # If all=false, generate UTXOs only for these wallets
      - "Node1"
      - "User1"
    nmin: 5               # Minimum number of UTXOs per wallet
    nmax: 10              # Maximum number of UTXOs
    vmin: 5               # Minimum amount per generated UTXO
    vmax: 10              # Maximum amount per generated UTXO

blockchain:
  nodes: 5                # Number of nodes (each node has a wallet)
  wallets: 10             # Number of additional wallets
  coinbase_start: 1000000 # Initial coinbase amount
  reward: 5               # Mining reward
  mining_diff: "20"       # Mining difficulty; higher means more difficult
  nonce_max: 2147483647   # Maximum nonce to limit mining time; 0 means no limit
```

## Launch

1. Clone the repository:

```bash
git clone https://github.com/IKarasev/bcatt.git
```

2. Navigate to the root project directory `bcatt`

3. Download Go dependencies:

```bash
go mod tidy
```

4. And run it:

```bash
go run ./cmd/main.go
```

By default, the web interface will be available at  
http://127.0.0.1:8080

Alternatively, download a precompiled executable for your OS from:  
https://github.com/IKarasev/bcatt/releases/tag/FirstRelease

## UI

### Start screen
![alt text](https://github.com/IKarasev/bcatt/releases/download/Images/start_screen.png) 

### Tab "Блоки"
![alt text](https://github.com/IKarasev/bcatt/releases/download/Images/blocks_1.png)
![alt text](https://github.com/IKarasev/bcatt/releases/download/Images/blocks_2.png)
![alt text](https://github.com/IKarasev/bcatt/releases/download/Images/blocks_3.png)

### Tab "Кошелек"
![alt text](https://github.com/IKarasev/bcatt/releases/download/Images/wallet_1.png) 
![alt text](https://github.com/IKarasev/bcatt/releases/download/Images/wallet_2.png) 

### Tab "Злодей"
![alt text](https://github.com/IKarasev/bcatt/releases/download/Images/evil_1.png) 
![alt text](https://github.com/IKarasev/bcatt/releases/download/Images/evil_2.png) 

### Visualizations
![alt text](https://github.com/IKarasev/bcatt/releases/download/Images/visual_2.png)
![alt text](https://github.com/IKarasev/bcatt/releases/download/Images/fork_3.png) 

