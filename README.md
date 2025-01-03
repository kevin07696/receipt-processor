# Receipt Processor
A challenge given by fetch-rewards

## Resource Description
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
    - Score stores a uint16 using 2 bytes.
    - Each ListNode store requires a left and right pointer. In a 64 bit architecture that is 8bits each.
    - Each ListNode also stores the UUID as a key and Score as a value in a `list.Entry{}` that points to the ListNode
    - ListNode = 36 bytes + 2 bytes + 8 bytes * 3 = 62 bytes
    - Map stores the uuid as a key and the listNode pointer as the value
    - Map = 36 bytes + 8 bytes = 44 bytes
    - Total per set transaction =  106 bytes
    - My allocated memory is 50MB: 52,428,800 bytes / 106 bytes ≈ 494,611 elements
- **Cons to `BigCache`:** 
  - I don't want to implement sharding because of its complexity and memory overhead.
  - Fractioning your data also slows down each process since it needs to implement a hash to find which shard the data belongs to.
  - Maintenance requires fine tuning parameters for sharding increasing complexity.
- **Pros to `BigCache`:**
  - At scale it can properly utilize sharding and parallel transaction greatly improving performance.
  - It also support expiration, so if the feature is utilized it can be another tool alongside eviction policy to keep the memory lean.
- **Pros to `Slice`**
  - I initially start with a slice, in which the id would just be the index and the element would hold a uint16 for score. It would be the more efficient in terms of space and time complexity.
- **Cons to `Slice`**
  - It is not built-in to be concurrent
  - It also would work poorly with uuids. UUIDs can convert into a uint for indexing, but it isn't guaranteed to be sequential.
