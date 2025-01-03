# Receipt Processor
A challenge given by fetch-rewards

## Resource Description
| Features                                           | lruCache | sync.Map | map | BigCache |
|----------------------------------------------------|----------|----------|-----|----------|
| **Eviction Policy (Allocated Memory Management)**  | Yes      | No       | No  | Yes      |
| **Concurrency (Multi_Thread Support)**             | Yes      | Yes      | No  | Yes      |
| **TTL (Expiration Support)**                       | No       | No       | No  | Yes      |
| **Sharding**                                       | No       | No       | No  | Yes      |

### Explanations:
- **Eviction**: This feature prevents a data store from ballooning in memory and using up all its allocated resources.
- **Concurrency**: This feature supports multiple threads accessing the cache simultaneously
- **TTL (Time-to-Live)**: This feature is an expiration for your cached item. Once the item's time has expired, the cache will remove the item.
- **Sharding**: This feature involves dividing the cache into smaller, manageable segments (shards). Each shard can be managed independently, which helps in distributing the load and improving performance, especially in large-scale caching scenarios.
- 

### Notes:
- **lruCache**: Custom implementation that supports eviction policy and concurrency.
- **sync.Map**: Built-in Go package for concurrent map operations without eviction.
- **map**: Standard Go map, not safe for concurrent use.
- **BigCache**: Third-party library for large-scale caching with eviction and concurrency support.

### Recommendations:
- I will be using the `lruCache`, because I need support for eviction policy to prevent the service from overusing its allocated cpu and memory. It also uses a doubly linkedList to order the requests by least recently used (lru) and pruning the oldest.
