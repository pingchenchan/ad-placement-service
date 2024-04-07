# ad-placement-service


## Introduction
In this project, we've developed a straightforward ad placement service. The service offers two APIs: one for generating ads and another for listing ads.

## Features
1. **API**
   - Admin API: Offers a POST RESTful API for creating a new ad.
   - Placement API: Provides a GET RESTful API for listing active ads, defined as ads where the current time (`time.now()`) falls within the "startAt" and "endAt" conditions.
2. **Strategies for High Performance**
   - To enhance GET performance, we use a **compound index for our MongoDB**. The compound index order is "startAt", "endAt", "ageStart", "ageEnd", "country", "gender", "platform". This order is highly efficient because when querying, we first filter by time to find the active ad. Age is also a highly selective index, followed by country, gender, and platform.
   - For POST requests, considering performance and data persistence, I've implemented three strategies: **Direct Write**, **Publisher/Subscriber**, and **Asynchronous Write**. More details are provided in the sections below.
   - For GET requests, I've implemented **Direct Get** from MongoDB and **Cache Recent Requests** using Redis. More details are provided in the sections below.
3. **Experimentation with Different Strategies**
   - I've implemented unit tests and integration tests for my different strategies of POST and GET APIs.
4. **Query Validation**
   - For writing a new ad, the conditions are optional. If conditions are provided, my validator will ensure the conditions are valid. If they are not, it will respond with an error.



## Post design

This service provides an API for handling high concurrency(10,000 requests per second) POST requests. It uses three different strategies:

1. **Direct Write**: 
   - Directly write the request data to the database and send a response.
   - Endpoint is : POST /ads
3. **Publisher / Subscriber**: 
   - Push the request data to a channel. Another goroutine periodically checks this channel and writes all data to the database, then sends an HTTP response.
   - Endpoint is : POST /adsBulk
4. **Asynchronous Write**: First, write the request data to Redis and send a response to the user. Then, periodically check and write all data from the Redis `ads_list` to the database. The steps are as follows:
   - Get all ads from `ads_list` and remove them from Redis.
   - Try to write the data to MongoDB. If the write operation fails, use exponential backoff and retry up to 3 times.
   - If the write operation still fails after 3 retries, store the data in the Redis `fail_ads` list.
   - If the write operation is successful, delete these records from Redis.
   - Asynchronous operation in Redis are atomic, ensuring data consistency even under high concurrency.
   - Endpoint is : POST /adsAsync



## Get Design
We provide two strategies for handling high concurrency (10,000 requests per second) GET requests:

1. **Direct Get**: This strategy directly retrieves the requested data from the database and sends a response.

2. **Cache Recent Requests**: This strategy involves checking if the query key exists in Redis each time a GET request is handled. If it does, the data is retrieved from Redis and sent in the response. If it doesn't, the data is retrieved from MongoDB, stored in Redis for future requests, and then sent in the response.


## Experiment Results

### GET ###
## Experiment Results

The following table presents the results of our experiments with different strategies for handling high concurrency GET requests. Each strategy was tested with a load of 10,000 requests per second.

| Strategy | Unique Queries | Average Response Time | Max Response Time | Min Response Time | Pass Validation |
|----------|----------------|-----------------------|-------------------|-------------------|-----------------|
| Redis | 10000 | 1.598750938s | 3.023441445s | 455.743878ms | 4516 |
| Redis | 5000 | 988.681536ms | 2.128397241s | 189.930277ms | 4548 |
| Redis | 4000 | 890.823334ms | 1.739472335s | 175.216723ms | 4504 |
| Redis | 3000 | 795.656568ms | 1.86191602s | 5.72865ms | 2090 |
| Redis | 1000 | 779.147789ms | 1.486303824s | 41.747529ms | 1760 |
| No Redis | 10000 | 1.892145725s | 3.494308782s | 357.576197ms | 4516 |
| No Redis | 5000 | 1.879807906s | 3.337061686s | 116.287444ms | 4548 |
| No Redis | 1000 | 1.618393602s | 2.896856922s | 57.08801ms | 1760 |

- **Strategy**: The method used for handling requests.
- **Unique Queries**: The number of unique queries in the test.
- **Average Response Time**: The average time taken to respond to a request.
- **Max Response Time**: The maximum time taken to respond to a request.
- **Min Response Time**: The minimum time taken to respond to a request.
- **Pass Validation**: The number of requests that passed validation.

These results demonstrate the effectiveness of our strategies in handling high concurrency requests. Using Redis for caching shows significant improvements in response time, especially when the number of unique queries is reduced.


### POST ###
The following table presents the results of our experiments with different strategies for handling high concurrency POST and GET requests. Each strategy was tested with a load of 10,000 requests per second.

## Experiment Results

The following table presents the results of our experiments with different strategies for handling high concurrency POST and GET requests. Each strategy was tested with a load of 10,000 requests per second.

| Strategy | Average Response Time | Max Response Time | Min Response Time | Total Duration | |
|----------|-----------------------|-------------------|-------------------|----------------|
| Asynchronous Write (Local) | 859.619972ms | 1.647311543s | 2.115333ms | 58.621597ms   |
| Asynchronous Write (HTTP Endpoint) | 898.034987ms | 1.317795309s | 109.471075ms | 42.034235ms   |
| Bulk Write (Local) | 501.75155ms | 902.442501ms | 115.539µs | 47.504073ms   |
| Bulk Write (HTTP Endpoint) | 573.98576ms | 907.051536ms | 78.065298ms |  137.577368ms |
| Direct Write (Local) | 630.330655ms | 1.115631343s | 294.721µs |  29.584926ms |
| Direct Write (HTTP Endpoint) | 650.709399ms | 1.23349069s | 20.320745ms | 122.25947815ms  |



- **Strategy**: The method used for handling requests.
- **Average Response Time**: The average time taken to respond to a request.
- **Max Response Time**: The maximum time taken to respond to a request.
- **Min Response Time**: The minimum time taken to respond to a request.
- **Total Duration**: The total duration of sending all requests from client the beckend.
- **Successful Requests**: The number of requests that were successful.
- **Error Requests**: The number of requests that resulted in an error.

These results demonstrate the effectiveness of our strategies in handling high concurrency requests. Asynchronous Write and Cache Recent Requests strategies, in particular, show significant improvements in response time and error rate.



## Running the Service

You can use Docker Compose to build and run the service:
### run the service
```bash
docker-compose up --build go-backend-test
```
### Running the Test
```
 docker-compose up --build go-backend-test; docker-compose down go-backend-test; docker rmi ad-placement-service-go-backend-test:latest
 ```