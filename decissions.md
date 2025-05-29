# Non Functional Dependencies

* [X] Micro Service
* [X] Error Handling/Input Validation
* [ ] Corn Jobs for expiring subscriptions
* [ ] Retry Mechanism
* [ ] Low Latency
* [X] Encryption
* [X] Redis caching

# Functional Dependencies

* [ ] User sub management
* [ ] Manage sub
* [ ] Retrive sub details
* [ ] Upgrade sub
* [ ] Cancel sub

# Tech Stack

1. Lang : Go
2. Backend services : Fiber
3. Database : Postgress ( Dockerized )
4. Databse ORM : Gorm
5. Auth : jwt
6. Secret management : godotenv
7. Validation : go-playground/validator
8. Cron jobs : robfig/cron
9. Logging : logrus
10. Redis for caching : go-redis/redis
11. Database migration : golang-migrate/migrate
12. Documentation : swaggo/swag (reads go comments and generates Swagger JSON)
13. Swagger UI : swaggo/gin-swagger

# Architecture
