# A project as an attempt to lab with Zookeeper

## How to run the code

### App
- Browse to /z-distribution
- Copy the .env.example to .env
- Run `go run .`
- The CMD app would accept these types of input:
    - GET: get the current counter
    - INC: increase the current counter with 1
    - SLEEPINC: increase the current counter with 1, but the process gonna sleep 10 seconds (to force race condition occurred)
- As many CMDs as possible to open, these CMDs would share the counter as the distribution lock
- This project is an attempt to lab with this lock 


### Zookeepers
- Browse to /zookeepers
- Run `docker-compose up` run zookeeper cluster (with 3 node)