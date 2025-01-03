# Receipt Processor
A challenge given by fetch-rewards

## Resources
- [ ] Golang 
- [ ] Docker (optional)
- [ ] Docker Compose (optional)

## Endpoints
| Method | Path                   | Request Body                      | Response Body                      |
|--------|------------------------|-----------------------------------|------------------------------------|
| POST   | /receipts/process      | JSON body with `Receipt` object   | JSON body with `UUID`              |
| GET    | /receipts/{id}/points  | URL Path Parameter `ID` string    | JSON body with `Points` (int64)    |
| GET    | /health                | None                              | JSON body with status `OK`         |

## Models

### Receipt
| Fields         | Type     | JSON          | Regex Pattern                                             |
|----------------|----------|---------------|-----------------------------------------------------------|
| Retailer       | string   | retailer      | `^[\w\s\-&]+$`                                            |
| PurchaseDate   | string   | purchaseDate  | `^[0-9]{4}-(0[1-9]\|1[0-2])-(0[1-9]\|[12][0-9]\|3[01])$`  |
| PurchaseTime   | string   | purchaseTime  | `^(0[0-9]\|1[0-9]\|2[0-3]):([0-5][0-9])$`                 |
| Items          | []Item   | items         |                                                           |
| Total          | string   | total         | `^\d+\.\d{2}$`                                            |

### Item
| Fields             | Type     | JSON               | Regex Pattern     |
|--------------------|----------|--------------------|-------------------|
| ShortDescription   | string   | shortDescription   | `^[\w\s\-]+$`     |
| Price              | string   | total              | `^\d+\.\d{2}$`    |

## Request Examples

### Method=`POST` Path=`/receipts/process`
```json
{
  "retailer": "Target",
  "purchaseDate": "2022-01-01",
  "purchaseTime": "13:01",
  "items": [
    {
      "shortDescription": "Mountain Dew 12PK",
      "price": "6.49"
    },{
      "shortDescription": "Emils Cheese Pizza",
      "price": "12.25"
    },{
      "shortDescription": "Knorr Creamy Chicken",
      "price": "1.26"
    },{
      "shortDescription": "Doritos Nacho Cheese",
      "price": "3.35"
    },{
      "shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
      "price": "12.00"
    }
  ],
  "total": "35.35"
}
```
#### Debug Level Logs

```
receipt_processor  | time=2025-01-03T20:27:53.236Z level=INFO msg="Method POST, Path: /receipts/process"
receipt_processor  | time=2025-01-03T20:27:53.237Z level=DEBUG msg="6 points - retailer name has 6 characters" RequestID=cceb6b85-4c6c-4266-80d7-cbc0e1b9a4a3
receipt_processor  | time=2025-01-03T20:27:53.237Z level=DEBUG msg="5 items (2 batches @ 5.00 points each)" RequestID=cceb6b85-4c6c-4266-80d7-cbc0e1b9a4a3
receipt_processor  | time=2025-01-03T20:27:53.237Z level=DEBUG msg="6 points - purchase day is odd" RequestID=cceb6b85-4c6c-4266-80d7-cbc0e1b9a4a3
receipt_processor  | time=2025-01-03T20:27:53.237Z level=DEBUG msg="3 Points - \"Emils Cheese Pizza\" is 18 characters (a multiple of 3) item price of 12.25 * 0.20 = 2.45 is rounded up is 3" RequestID=cceb6b85-4c6c-4266-80d7-cbc0e1b9a4a3
receipt_processor  | time=2025-01-03T20:27:53.237Z level=DEBUG msg="3 Points - \"Klarbrunn 12-PK 12 FL OZ\" is 24 characters (a multiple of 3) item price of 12.00 * 0.20 = 2.40 is rounded up is 3" RequestID=cceb6b85-4c6c-4266-80d7-cbc0e1b9a4a3
receipt_processor  | time=2025-01-03T20:27:53.237Z level=INFO msg="Total Points: 28" RequestID=cceb6b85-4c6c-4266-80d7-cbc0e1b9a4a3
```
#### Response
```json
{
  "ID": "edef5a0a-7dc5-4b56-97a1-b0007f3d8355"
}
```

### Method=`GET` Path=`/receipts/{id}/points`
```
GET http://localhost:3000/receipts/edef5a0a-7dc5-4b56-97a1-b0007f3d8355/points
```
#### Debug Level Logs
```
receipt_processor  | time=2025-01-03T20:36:08.572Z level=INFO msg="Method GET, Path: /receipts/edef5a0a-7dc5-4b56-97a1-b0007f3d8355/points"
```
### Method=`GET` Path=`/health`
```
GET http://localhost:3000/health
```
#### Response
```
OK
```

## Cache
| Features                                           | BigCache | lruCache | sync.Map | map |
|----------------------------------------------------|----------|----------|----------|-----|
| **Eviction Policy (Allocated Memory Management)**  | Yes      | Yes      | No       | No  |
| **TTL (Expiration Support)**                       | Yes      | No       | No       | No  |
| **Sharding**                                       | Yes      | No       | No       | No  |
| **Concurrency (Multi-Threading)**                  | Yes      | Yes      | Yes      | No  |
| **Parallelism (Multi-Threading)**                  | Yes      | No       | No       | No  |



### Features:
- **Eviction Policy**: This feature prevents a data store from ballooning in memory and using up all its allocated resources.
- **Concurrency**: This feature supports multiple threads accessing the cache simultaneously by utilizing context switching and/or synchronization mechanisms like mutex. 
- **TTL (Time-to-Live)**: This feature is an expiration for your cached item. Once the item's time has expired, the cache will remove the item.
- **Sharding**: This feature involves dividing the cache into smaller, manageable segments (shards). Each shard can be managed independently, which helps in distributing the load and improving performance, especially in large-scale caching scenarios.
- **Parallism**: This feature supports accessing the cache in parallel threads. BigCache uses sharding to distribute incoming data across multiple independent segments (shards). Each shard can handle requests independently, allowing for parallel processing of cache operations. This design minimizes contention and improves performance, especially in high-concurrency environments.

### Caches:
- **lruCache**: Custom implementation that supports eviction policy and concurrency.
- **sync.Map**: Built-in Go package for concurrent map operations without eviction.
- **map**: Standard Go map, not safe for concurrent use.
- **BigCache**: Third-party library for large-scale caching with eviction and concurrency support.

### Requirements:
- I need support for eviction policy to prevent the service from overusing its allocated cpu and memory.
- I need support for concurrency to prevent race conditions between threads

### My Choice
- I chose to use `lruCache` for its simplicity.
- **Cons to `lruCache`:**
  - At maximum capacity it wil reach a bottleneck. Each set operation afterwards will need to run an eviction of it least recently used data.
- **Pro to `lruCache`:**
  - It is simple to implement. Only requiring a capacity of elements. To use it calculate the memory required to store each element:
    - UUID string at 36 bytes
    - Score stores a int64 using 8 bytes.
    - Each ListNode store requires a left and right pointer. In a 64 bit architecture that is 8 bytes each.
    - Each ListNode also stores the UUID as a key and Score as a value in a `list.Entry{}` that points to the ListNode
    - ListNode = 36 bytes + 8 bytes + 8 bytes * 3 = 68 bytes
    - Map stores the uuid as a key and the listNode pointer as the value
    - Map = 36 bytes + 8 bytes = 44 bytes
    - Total per set transaction =  112 bytes
    - My allocated memory is 50MB: 52,428,800 bytes / 112 bytes ≈ 468114 elements
- **Cons to `BigCache`:** 
  - I don't want to implement sharding because of its complexity and memory overhead.
  - Fractioning your data also slows down each process since it needs to implement a hash to find which shard the data belongs to.
  - Maintenance requires fine tuning parameters for sharding increasing complexity.
- **Pros to `BigCache`:**
  - At scale it can properly utilize sharding and parallel transaction greatly improving performance.
  - It also support expiration, so if the feature is utilized it can be another tool alongside eviction policy to keep the memory lean.
- **Pros to `Slice`**
  - I initially start with a slice, in which the id would just be the index and the element would hold a int64 for score. It would be the more efficient in terms of space and time complexity.
- **Cons to `Slice`**
  - It is not built-in to be concurrent
  - It also would work poorly with uuids. UUIDs can convert into a uint for indexing, but it isn't guaranteed to be sequential.
