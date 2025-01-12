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

## Installation

1. **Clone the Repository:**
   ```bash
   git clone https://github.com/yourusername/your-api-repo.git
   cd your-api-repo
2. **Build Docker Image:**
   ```bash
   docker build -t receipt-processor
   ```
4. **Run Docker Container:**
  ```bash
  docker run -d -p 3000:8080 receipt-processor
  ```
** Optionally you can also run Docker Compose and use the health check to monitor container health
1. **Build containers with Docker Compose:**
 ```bash
 docker-compose build
 ```
2. **Run containers with Docker Compose:**
```bash
docker-compose up
```
3. **To stop and remove the containers, networks, and volumes**
```bash
docker-compose down
```
## Environment Variables
### Config Variables
1. HOST_PORT=3000
   - Definition: Host port specifies the port the application is listening to on the host machine
2. APP_PORT=8080
   - Definition: App port specifies the port the application is listening to inside the container
   - Example: If the application's port specification is `3000:8080`, if you are making calls inside the container you are sending requests to `8080`, but from your local machine you would be calling `http://localhost:3000`  
3. APP_ENV=DEVELOPMENT
   - Usage: This variable is to toggle environmental settings.
   - Example: In this application, logging level and logging handler is determined by environment. In development, I am running in debug level with a plain text log handler. In production it is info level with a JSON log handler 
4. CACHE_CAP=200000
  - Definition: Cache's capacity is based on the number elements.
  - Usage: Determine the space allocated for the application / the space allocated for one item to determine the capacity*2. This way you only use half the allocated memory. 

### Multiplier Variables
1. MULT_RECEIPT=1
   - Definition: Multiplier for receipt name score
2. MULT_ROUND_TOTAL=50
   - Definition: Multiplier for round total score
3. MULT_DIVISIBLE_TOTAL=25
   - Definition: Multiplier for score, in which the total is divisible by 0.25
4. MULT_ITEMS=5
   - Definition: Multiplier for each pair of items
5. MULT_DESCRIPTION=0.2
   - Definition: Multplier for short description
6. MULT_PURCHASE_TIME=10
   - Definition: Multiplier for purchase time  
7. MULT_PURCHASE_DATE=6
    - Definition: Multiplier for purchase date

### Score Rule Variables
1. START_TIME=14:00
   - Definition: Specify the purchase start time
2. END_TIME=16:00
   - Definition: Specify the purchase end time
   - Usage: Start time needs to be before end time
3. TOTAL_MULTIPLE=0.25
   - Definition: Total's divisible conditional. The default is 0.25.
4. ITEMS_MULTIPLE=2
   - Definition: Items' divisible conditional. The default is each pair gets a point. Thus, round down.
5. DESCRIPTION_MULTIPLE=3
   - Definition: Description length divisible condtional. Challenge specifies to round up.

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
** The process method is now idempotent. The id is generated from the whole receipt request body. I tested that the receipt will generate the same uuid. Thus, the validation and point calculation will not run and not be logged for repeated requests.

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

## Error Log Book

### Fatal Errors
Since I didn't see a response scenario for 500 errors in the api.yaml, I am trying to set a system that immediately exits when the container/application instance is unhealthy. Then, it spins up a healthy container. This was popularized by Erlang where sometimes older instances can become unhealthy. Since Erlang applications are known for their persistence, it is appropriate for them to spin up a new fresh container. However, a lot of these errors might be better as `Bad Requests`, but the issue isn't just the request, but the possible missing validation that allowed it to occur.

1. `Error loading .env file`
   - Environmental variables are required. I would check that the .env file is at the root to access. Also, check the same for the docker container. If the file is not placed at the root with main.go, then main() won't be able to find env. 
2. `Server failed to start: %v`
   - Server can fail for a number of reasons:
      - Check that it is not the middleware.
      - Check that the host and container ports are available
3. `Error parsing %s: %v`
   - Check environmental variables to make sure the types can match config
4. `Unsupported type for environment variable %s"`
   - This can happen if you add/update variables, but don't update the config parsing method to support that new type
5. `Failed to store value: %v: %v`
   - This log is in BigCache's set method. This is can happen if the allocated memory can not fit an element. To use the api, you can allocate for more memory
6. `Failed to parse %s, %s. Check validation: %v`
   - This I would usually run as a bad request error; however, the api.yaml did not have a 500 error type.
   - Also, it should not happen, so if it does then that means that sample should be recorded to support the validation error
7. `Failed to marshal response: %v`
   - This should not happen since the response is created by the application with zero data from request. Thus, if you get this message I would check the logged response to see why it is invalid and then look into the logic used to wrap the response.
