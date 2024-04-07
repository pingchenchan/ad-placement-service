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
3. 




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