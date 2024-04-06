## this is the dcard assignment
please download the zip files on redis and worker branch respectively,and then unzip:

environment: 
**docker compose version**: Docker Compose version v2.18.1 on Mac
If you use Windows or Linux, you can use Docker Compose version v2.26.1

- you can simply download the zip files on the tag 0.0.1 or tag 0.0.2

1. first, run **docker build -t dcard:latest .** in the folder corresponding to the redis folder
2. second, run **docker build -t dcard-background:latest  .** in the folder corresponding to the worker folder
3. And finally run **docker compose up -d** in the folder corresponding to the redis folder


### Test this api: 

- usage: 
For the get request endpoint /api/v1/ad, you can use the following curl command: 

> curl http://localhost:8081/api/v1/ad?offset=0&limit=2&age=24&gender=F&country=TW

or 

> curl http://localhost:8081/api/v1/ad?offset=1&limit=3&age=24&gender=F&country=TW&platform=ios

#### only limit parameter is necessary, the others are optional

For post request, you can create a post request with the body like this: 
> curl -X POST -H "Content-Type: application/json" -d '{
  "title": "AD3086",
  "startAt": "2024-01-29T04:23:10.000Z",
  "endAt": "2024-03-31T09:25:30.000Z",
  "conditions": {
    "ageStart": 34,
    "ageEnd": 55,
    "country": [
      "TW",
      "JP"
    ],
    "gender": "F",
    "platform": [
      "ios"
    ]
  }
}' http://localhost:8081/api/v1/ad

* It's the same as the example listed on the assignment

### The ways that I implement this project:

1. use redis pool to restrict the maximum number of connections: 
> avoid the pc from running out of conneections and memory resources 
2. set the maximum idle connections in redis pool: 
> let the resource can be utilized more efficiently
3. set the ttl for key-value response which will be sent to client: 
> let some responses on least frequent used requests be erased from redis cache
4. If client sent too much get requests to api server, originally, it will be handled by gin-server.But, when the number of requests is over the maximum requests(by default 1000, I set in service/advertise.go.Of course, you can tune it if you want), the api will send the excedded requests to the worker queue to handle
5. Before implementing the worker-client model for this project, I just restrict the maximum number of go routimes which gin-handler will create for requests. This implementation is under the v0.0.0 tag

### performance

* I use the "wrk" utility to test the performance for api endpoints, and you can check api performance by your own tools. I list the outputs of api performance on the redis branch

