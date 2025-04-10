# Hệ thống theo dõi đơn hàng logistics

Hệ thống theo dõi đơn hàng logistics là một ứng dụng sử dụng mô hình Event Sourcing và CQRS (Command Query Responsibility Segregation) để quản lý và theo dõi trạng thái của các đơn hàng trong quá trình vận chuyển.

## Mô hình kiến trúc

### Event Sourcing

Event Sourcing là một mô hình kiến trúc lưu trữ trạng thái của hệ thống dưới dạng một chuỗi các sự kiện (events). Mỗi thay đổi trạng thái được ghi lại như một sự kiện mới, thay vì cập nhật trực tiếp vào trạng thái hiện tại. Điều này cho phép:

- Lưu trữ lịch sử đầy đủ của mọi thay đổi
- Khôi phục trạng thái tại bất kỳ thời điểm nào trong quá khứ
- Tái phát (replay) các sự kiện để tạo các khung nhìn (view) khác nhau của dữ liệu

### CQRS (Command Query Responsibility Segregation)

CQRS là mô hình tách biệt thao tác đọc (queries) và thao tác ghi (commands) thành các thành phần riêng biệt:

- **Command**: Xử lý các thao tác thay đổi dữ liệu (tạo, cập nhật, xóa)
- **Query**: Xử lý các thao tác đọc dữ liệu, thường từ các khung nhìn (views) được tối ưu hóa cho việc truy vấn

Việc tách biệt này cho phép tối ưu hóa riêng cho từng loại thao tác và dễ dàng mở rộng hệ thống.

## Cấu trúc dự án

```
logistics/
├── cmd/
│   └── logistics/
│       └── main.go             # Điểm khởi chạy ứng dụng
├── internal/
│   ├── command/                # Xử lý các lệnh (thao tác ghi)
│   │   ├── commands.go         # Định nghĩa các lệnh
│   │   └── handlers.go         # Xử lý các lệnh
│   ├── query/                  # Xử lý các truy vấn (thao tác đọc)
│   │   ├── queries.go          # Định nghĩa các truy vấn
│   │   └── handlers.go         # Xử lý các truy vấn
│   ├── domain/                 # Mô hình domain và logic nghiệp vụ
│   │   ├── order.go            # Order aggregate
│   │   └── events.go           # Định nghĩa các sự kiện
│   ├── eventstore/             # Lưu trữ sự kiện
│   │   ├── store.go            # Interface của event store
│   │   └── postgres.go         # Triển khai với PostgreSQL
│   ├── projection/             # Xây dựng khung nhìn từ các sự kiện
│   │   ├── order.go            # Interface của projection
│   │   ├── postgres_order.go   # Triển khai với PostgreSQL
│   │   └── postgres_tracking.go # Projection cho theo dõi đơn hàng
│   └── server/                 # API server
│       ├── handlers.go         # HTTP handlers
│       └── routes.go           # Định nghĩa các route
└── pkg/
    └── eventbus/               # Event bus cho pub/sub
        └── bus.go              # Triển khai event bus
```

## Luồng dữ liệu

1. **Command Flow (Write)**:
   - Client gửi lệnh (command) tới API
   - Command Handler xác thực lệnh
   - Command Handler lấy trạng thái hiện tại từ Event Store (nếu cần)
   - Command Handler thực thi logic nghiệp vụ và tạo sự kiện mới
   - Event Store lưu trữ sự kiện vào cơ sở dữ liệu
   - Event Bus phát sự kiện cho các Projection

2. **Query Flow (Read)**:
   - Client gửi truy vấn (query) tới API
   - Query Handler xác thực truy vấn
   - Query Handler truy xuất dữ liệu từ khung nhìn đọc (read model)
   - Kết quả được trả về cho client

3. **Projection Flow**:
   - Projection lắng nghe các sự kiện từ Event Bus
   - Projection cập nhật khung nhìn đọc (read model) dựa trên sự kiện
   - Khung nhìn đọc được tối ưu hóa cho truy vấn

## Mô hình dữ liệu

### Event Store

```sql
CREATE TABLE events (
    id          VARCHAR(36) PRIMARY KEY,
    aggregate_id VARCHAR(36) NOT NULL,
    type        VARCHAR(50) NOT NULL,
    version     INTEGER NOT NULL,
    data        JSONB NOT NULL,
    metadata    JSONB,
    timestamp   BIGINT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_aggregate_id ON events (aggregate_id);
CREATE INDEX idx_events_type ON events (type);
CREATE INDEX idx_events_timestamp ON events (timestamp);
```

### Read Models

#### Orders

```sql
CREATE TABLE orders (
    id                VARCHAR(36) PRIMARY KEY,
    customer_id       VARCHAR(36) NOT NULL,
    tracking_number   VARCHAR(36) NOT NULL UNIQUE,
    status            VARCHAR(20) NOT NULL,
    origin_data       JSONB NOT NULL,
    destination_data  JSONB NOT NULL,
    current_location_data JSONB,
    items_data        JSONB NOT NULL,
    notes_data        JSONB NOT NULL,
    created_at        TIMESTAMP NOT NULL,
    updated_at        TIMESTAMP NOT NULL
);

CREATE INDEX idx_orders_customer_id ON orders (customer_id);
CREATE INDEX idx_orders_tracking_number ON orders (tracking_number);
CREATE INDEX idx_orders_status ON orders (status);
```

#### Tracking

```sql
CREATE TABLE tracking_info (
    id                VARCHAR(36) PRIMARY KEY,
    order_id          VARCHAR(36) NOT NULL,
    tracking_number   VARCHAR(36) NOT NULL UNIQUE,
    status            VARCHAR(20) NOT NULL,
    origin_data       JSONB NOT NULL,
    destination_data  JSONB NOT NULL,
    current_location_data JSONB,
    estimated_delivery TIMESTAMP,
    created_at        TIMESTAMP NOT NULL,
    updated_at        TIMESTAMP NOT NULL
);

CREATE TABLE tracking_updates (
    id                VARCHAR(36) PRIMARY KEY,
    tracking_id       VARCHAR(36) NOT NULL,
    timestamp         TIMESTAMP NOT NULL,
    status            VARCHAR(20),
    location_data     JSONB,
    message           TEXT NOT NULL
);

CREATE INDEX idx_tracking_order_id ON tracking_info (order_id);
CREATE INDEX idx_tracking_tracking_number ON tracking_info (tracking_number);
CREATE INDEX idx_tracking_updates_tracking_id ON tracking_updates (tracking_id);
CREATE INDEX idx_tracking_updates_timestamp ON tracking_updates (timestamp);
```

## API Endpoints

### Commands (Write)

- `POST /api/soa/v1/logistics/orders` - Tạo đơn hàng mới
- `PUT /api/soa/v1/logistics/orders/{id}/status` - Cập nhật trạng thái đơn hàng
- `POST /api/soa/v1/logistics/orders/{id}/cancel` - Hủy đơn hàng
- `POST /api/soa/v1/logistics/orders/{id}/notes` - Thêm ghi chú vào đơn hàng

### Queries (Read)

- `GET /api/soa/v1/logistics/orders` - Lấy danh sách đơn hàng
- `GET /api/soa/v1/logistics/orders/{id}` - Lấy chi tiết đơn hàng
- `GET /api/soa/v1/logistics/orders/{id}/history` - Lấy lịch sử đơn hàng
- `GET /api/soa/v1/logistics/orders/tracking/{tracking_number}` - Lấy đơn hàng theo số theo dõi
- `GET /api/soa/v1/logistics/tracking/{tracking_number}` - Lấy thông tin theo dõi đơn hàng

## Lợi ích của kiến trúc Event Sourcing và CQRS

1. **Lịch sử đầy đủ**: Lưu trữ mọi thay đổi trạng thái giúp kiểm tra, audit và hiểu rõ quá trình diễn ra.
2. **Khả năng mở rộng**: Tách biệt đọc/ghi cho phép mở rộng độc lập các phần của hệ thống.
3. **Tính linh hoạt**: Dễ dàng thêm các khung nhìn đọc mới mà không ảnh hưởng đến logic nghiệp vụ cốt lõi.
4. **Tính nhất quán**: Event Sourcing đảm bảo tính nhất quán cao trong hệ thống.
5. **Hỗ trợ debugging**: Dễ dàng tái hiện các vấn đề bằng cách xem lại chuỗi sự kiện.

## Thách thức và giải pháp

1. **Độ phức tạp**: Kiến trúc này phức tạp hơn so với CRUD truyền thống, nhưng lợi ích lâu dài đáng giá.
2. **Eventual Consistency**: CQRS thường sử dụng mô hình eventual consistency, cần thiết kế UI để xử lý điều này.
3. **Learning Curve**: Đội phát triển cần thời gian để làm quen với mô hình này.
4. **Quản lý schema**: Cần chiến lược để xử lý thay đổi trong cấu trúc sự kiện theo thời gian.

## Kết luận

Kiến trúc Event Sourcing và CQRS là lựa chọn phù hợp cho hệ thống theo dõi đơn hàng logistics, nơi lịch sử thay đổi trạng thái rất quan trọng và các mẫu truy vấn có thể khác biệt đáng kể so với cấu trúc lưu trữ dữ liệu. Mặc dù có độ phức tạp ban đầu cao hơn, nhưng lợi ích về tính linh hoạt, khả năng mở rộng và tính nhất quán sẽ mang lại giá trị lớn khi hệ thống phát triển theo thời gian.
# ENV

```
APP_ENV=dev
BASE_PATH=/api/soa/v1/

SERVER_PORT=80

DB_DRIVER=
DB_HOST=
DB_PORT=
DB_USER=
DB_PASS=
DB_NAME=

REDIS_HOST=
REDIS_PORT=
REDIS_PASS=
REDIS_INDEX=
REDIS_CLUSTER=

```

- Need Redis to Incr, Decr statistics

# Swagger

- [swagger.yaml](swagger.yaml)
- Can access to http//{host}{BASE_PATH}/doc => to view

![img-doc.png](img-doc.png)

# Source code 

- tools: 
    - go-kit https://gokit.io/
    - bun https://bun.uptrace.dev/
    - geoip2-golang https://github.com/oschwald/geoip2-golang ([GeoLite2-City.mmdb](GeoLite2-City.mmdb))
- Structure
  - [cmd](cmd)
    - [cmd.go](cmd%2Fcmd.go): migration command line manually
    - [main.go](cmd%2Fmain.go): main app
  - [cfg](cfg): config
  - [internal](internal)
    - [kit](internal%2Fkit): go-kit directory
      - [endpoints](internal%2Fkit%2Fendpoints): go-kit endpoint (can mapping as controller)
      - [services](internal%2Fkit%2Fservices): go-kit service (biz logic)
      - [transports](internal%2Fkit%2Ftransports): go-kit transport (can be https/gRPC)
    - [models](internal%2Fmodels): database models
    - [transforms](internal%2Ftransforms): mapping request
  - [migrations](migrations): migration file version
  - [pkgs](pkgs): external func
  - [server](server): init app