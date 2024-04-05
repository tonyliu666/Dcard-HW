## this is the HW for dcard

1. first, run docker build -t dcard:latest . on the redis branch
2. second, run docker build -t dcard-background:latest  . on the worker branch 
3. And finall run docker compose up -d on the redis branch


* Test this api: 

- usage: 
For the get request endpoint /api/v1/ad, you can use the following curl command: 

> curl http://localhost:8081/api/v1/ad?offset=1&limit=1&age=32&gender=F&country=TW
> curl http://localhost:8081/api/v1/ad?offset=1&limit=3&age=24&gender=F&country=TW&platform=ios

#### only limit parameter is necessary, the others are optional

For post request, you can create a post request with the body like this: 
> curl -X POST -H "Content-Type: application/json" \
"http://localhost:8081/api/v1/ad" \ --data '{
"title" "AD 55",
"startAt" "2023-12-10T03:00:00.000Z", "endAt" "2023-12-31T16:00:00.000Z", "conditions": {
{
"ageStart": 20,
"ageEnd": 30,
"country: ["TW", "JP"], "platform": ["android", "ios"]
} }
}'
* it's the same as the example listed on the assignment

### performance

> I use the "wrk" utility to test the performance for api endpoints, and you can check api performance by your own tools. I list the outputs of api performance on the redis branch

