# Usage
Open 6 terminal and in each one run one

p1 - go run server.go 
p2 - go run peer.go poisson.go localhost 8081 localhost:8082 false
p3 - go run peer.go poisson.go localhost 8082 localhost:8083 false
p4 - go run peer.go poisson.go localhost 8083 localhost:8084 false
p5 - go run peer.go poisson.go localhost 8084 localhost:8085 false
p6 - go run peer.go poisson.go localhost 8085 localhost:8081 true