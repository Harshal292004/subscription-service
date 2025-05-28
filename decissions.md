# Non Functional Dependencies

* [ ] Micro Service
* [ ] Error Handling/Input Validation
* [ ] Corn Jobs for expiring subscriptions
* [ ] Retry Mechanism
* [ ] Low Latency
* [ ] Encryption
* [ ] Redis caching
* [ ] RabitMQ queue event sstreaming

# Functional Dependencies

* [ ] User sub management
* [ ] Sub to various plans
* [ ] Manage sub
* [ ] Retrive sub details
* [ ] Upgrade sub
* [ ] Cancel sub

# Tech Stack 

1. Lang : Go
2. Backend services : Gin
3. Database : Postgress ( Dockerized )
4. Databse ORM : Gorm
5. Auth : jwt
6. Secret management : .goenv
7. Validation : go-playground/validator
8. Cron jobs : robfig/cron
9. Logging : logrus
10. Queue : RabitMQ (amqq)
11. Redis for caching : go-redis/redis
12. Testing : *testing framework
13. Database migration : golang-migrate/migrate
14. Documentation : swaggo/swag (reads go comments and generates Swagger JSON)
15. Swagger UI : swaggo/gin-swagger


# Architecture
