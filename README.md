## this is the dcard assignment
please download the zip files on redis and worker branch respectively,and then unzip:

environment: 
**docker compose version**: Docker Compose version v2.18.1 on Mac
If you use Windows or Linux, you can use Docker Compose version v2.26.1

1. first, run **docker build -t dcard:latest .** on the redis branch
2. second, run **docker build -t dcard-background:latest  .** on the worker branch 
3. And finally run **docker compose up -d** on the redis branch


### Test this api: 

- usage: 
For the get request endpoint /api/v1/ad, you can use the following curl command: 

> curl http://localhost:8081/api/v1/ad?offset=0&limit=2&age=24&gender=F&country=TW

or 

> curl http://localhost:8081/api/v1/ad?offset=1&limit=3&age=24&gender=F&country=TW&platform=ios

#### only limit parameter is necessary, the others are optional

For post request, you can create a post request with the body like this: 
> curl -X POST -H "Content-Type: application/json" -d '{
  "title": "AD256",
  "startAt": "2024-01-30T04:23:10.000Z",
  "endAt": "2024-03-31T09:25:30.000Z",
  "conditions": {
    "ageStart": 40,
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
}'
* It's the same as the example listed on the assignment

### performance

* I use the "wrk" utility to test the performance for api endpoints, and you can check api performance by your own tools. I list the outputs of api performance on the redis branch

