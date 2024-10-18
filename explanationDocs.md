
# Logistics Platform Architecture

## Table of Contents

1. [Overview](#overview)
2. [Directory Structure](#directory-structure-overview)
3. [Load Balancing](#load-balancing)
4. [Distributed Database Architecture](#distributed-database-architecture)
5. [Handling High-Concurrency with Go](#handling-high-concurrency-with-go)
6. [Real-Time Data Management](#real-time-data-management)
7. [Efficient Driver Matching Algorithm](#efficient-driver-matching-algorithm)
8. [Pricing Model](#pricing-model)
9. [Major Design Decisions and Trade-offs](#major-design-decisions-and-trade-offs)
10. [Future Enhancements for Scalability](#future-enhancements-for-scalability)
11. [Conclusion](#conclusion)

### Overview

The logistics platform is designed to be a highly scalable system capable of handling 10,000 concurrent requests per second, accommodating 50 million registered users and 100,000 registered drivers globally. The system connects users needing to transport goods with available drivers, providing real-time availability, pricing, and tracking. It also enables admins monitor available vehicles, drivers, and get insights. The implementation leverages Go's concurrency features, MongoDB's scalability, WebSockets for real-time communication, and modular architecture to achieve high performance and scalability.

### Directory Structure Overview
The project is organized into a well-structured directory hierarchy that promotes modularity, maintainability, and scalability. Below is the directory tree with explanations for each part:

```go
.
├── cmd
│   └── backend
│       └── main.go
├── configs
│   └── config.yaml
├── go.mod
├── go.sum
├── internal
│   ├── api
│   │   └── router.go
│   ├── handlers/
│   ├── messaging
│   │   ├── messaging.go
│   │   ├── nats_client.go
│   │   └── websocket_client.go
│   ├── models/
│   ├── repositories/
│   ├── services/
│   └── utils
│       ├── config.go
│       ├── db.go
│       ├── logger.go
│       └── middleware.go
├── pkg
│   ├── auth/
│   ├── scheduler/
│   └── websocket/
├── README.md
```

#### Top-Level Directories and Files
* **cmd/backend/main.go**: The entry point of the application. It initializes the server, connects to the database, sets up services and handlers, and starts the HTTP server.

* **configs/config.yaml**: Configuration file containing environment-specific settings like server address, MongoDB URI, JWT secret, and messaging configurations.


#### Internal Directory Overview

The `internal` directory contains the core business logic of the application, organized into several key subdirectories:

1. **api**: Defines HTTP routes and middleware using the Gin framework.

2. **handlers**: Contains HTTP request handlers for various functionalities (admin, booking, driver, user, websocket).

3. **messaging**: Implements messaging interfaces for NATS and WebSocket communication.

4. **models**: Defines data structures for core entities (admin, booking, driver, user, vehicle, pricing).

5. **repositories**: Manages data access layers for different entities.

6. **services**: Implements business logic for various functionalities.

7. **distance**: Provides distance calculation methods (Google Maps API, Haversine formula).

8. **utils**: Contains utility functions for configuration, database connections, logging, and middleware.

The `pkg` directory includes reusable packages:

- **auth**: JWT-based authentication.
- **scheduler**: Cron-like task scheduler.
- **websocket**: WebSocket connection management.

This structure promotes modularity and separation of concerns, facilitating maintainability and potential future microservices refactoring.

### **Load Balancing**

#### **a. Stateless Service Design**

- **Implementation**: The backend services are designed to be stateless, meaning each request is independent and does not rely on server-side session data. Authentication is managed using JWT (JSON Web Tokens), which encapsulates user information within the token itself.
  
- **Benefits**:
  - **Horizontal Scalability**: Statelessness allows the addition of more server instances without the complexity of synchronizing session data across servers.
  - **Fault Tolerance**: If one server instance fails, others can seamlessly handle incoming requests without service disruption.

#### **b. External Load Balancers**

- **Implementation**: Deploying the backend behind an external load balancer (e.g., **NGINX**, **HAProxy**, or cloud-based solutions like **AWS Elastic Load Balancer**) distributes incoming traffic evenly across multiple server instances.
  
- **Features**:
  - **Health Checks**: Regular monitoring of server health ensures that traffic is only directed to healthy instances, enhancing reliability.
  - **SSL Termination**: Offloading SSL processing to the load balancer reduces the computational load on backend servers.

- **Benefits**:
  - **Scalability**: Efficiently manages high traffic volumes by balancing the load across numerous servers.
  - **Redundancy**: Prevents single points of failure, ensuring continuous availability.

### **Distributed Database Architecture**

#### **a. MongoDB Sharding**

- **Implementation**: The backend utilizes **MongoDB** as its primary database, leveraging its **sharding** capabilities to distribute data across multiple servers.
  
- **Shard Key Selection**:
  - **Criteria**: Keys are chosen based on high-cardinality fields such as `user_id` or `booking_id` to ensure even data distribution.
  - **Example**: Using `user_id` as a shard key allows user-related queries to be distributed across shards, preventing hotspots.

- **Benefits**:
  - **Horizontal Scalability**: Sharding enables the database to handle increased load by adding more shards, thus distributing read and write operations.
  - **Performance Optimization**: Distributes data to reduce the load on any single shard, enhancing query performance.

#### **b. Replica Sets for High Availability**

- **Implementation**: **MongoDB Replica Sets** are employed to ensure data redundancy and high availability.
  
- **Features**:
  - **Automatic Failover**: In case the primary node fails, a secondary node is automatically promoted to primary, ensuring uninterrupted service.
  - **Data Redundancy**: Multiple copies of data across different nodes safeguard against data loss.

- **Benefits**:
  - **Fault Tolerance**: Maintains database availability despite individual node failures.
  - **Read Scalability**: Distributes read operations across replica members, reducing the load on the primary node.

#### **c. Geospatial Indexing**

- **Implementation**: Geospatial indexes are created on driver locations within MongoDB to facilitate efficient proximity-based queries.
  
- **Usage**:
  - **Driver Matching**: Enables rapid retrieval of nearby available drivers based on user pickup locations.
  
- **Benefits**:
  - **Optimized Queries**: Significantly reduces query latency for location-based searches, essential for real-time driver assignments.
  - **Scalability**: Maintains performance even as the number of drivers scales to 100,000.

### **Handling High-Concurrency with Go**

#### **a. Goroutines for Lightweight Concurrency**

- **Implementation**: The backend leverages Go’s **goroutines** to handle numerous concurrent requests efficiently.
  
- **Benefits**:
  - **High Throughput**: Efficiently manages 10,000 concurrent requests per second by utilizing Go’s optimized concurrency model.
  - **Resource Efficiency**: Maximizes CPU and memory usage, ensuring optimal performance under heavy load.

#### **b. Channel-Based Communication**

- **Implementation**: **Channels** in Go are used for safe and synchronized communication between goroutines, particularly within the `WebSocketHub`.
  
- **Usage**:
  - **Message Broadcasting**: Channels handle the broadcasting of messages to connected WebSocket clients without race conditions.
  
- **Benefits**:
  - **Thread Safety**: Prevents data races and ensures consistent message delivery.
  - **Scalability**: Supports the management of thousands of WebSocket connections seamlessly.

### **Real-Time Data Management**

#### **a. WebSocketHub for Real-Time Communication**

- **Implementation**: A centralized `WebSocketHub` manages all active WebSocket connections, facilitating real-time updates such as driver locations and booking statuses.
  
- **Features**:
  - **Connection Management**: Handles registration, unregistration, and broadcasting of messages to appropriate clients (users or admins).
  - **Broadcast Mechanism**: Efficiently distributes messages to relevant clients using channels, ensuring timely delivery.

- **Benefits**:
  - **Low Latency**: Enables instant transmission of real-time data, crucial for tracking and status updates.
  - **Scalability**: Designed to handle thousands of concurrent WebSocket connections without degrading performance.

#### **b. Messaging Systems Integration**

- **Implementation**: The backend supports both **WebSockets** and **NATS** (optional) for messaging, configurable via `config.yaml`.
  
- **Features**:
  - **WebSocketClient**: Facilitates direct real-time communication with clients.
  - **NATSClient**: Provides a scalable, distributed messaging solution for inter-service communication and event handling.
  
- **Benefits**:
  - **Flexibility**: Allows the system to choose the most suitable messaging mechanism based on deployment needs.
  - **Scalability**: NATS can handle high message volumes with low latency, supporting the platform’s scalability requirements.

### **Efficient Driver Matching Algorithm**

#### **a. Geospatial Queries**

- **Implementation**: Utilizes MongoDB’s geospatial indexes to perform proximity-based searches for available drivers matching the required vehicle type.
  
- **Process**:
  - **Query Optimization**: Leverages `$near` queries to find drivers within a specified radius, ensuring quick and efficient matching.
  
- **Benefits**:
  - **Performance**: Rapidly retrieves relevant drivers, essential for maintaining low response times under high traffic.
  - **Scalability**: Efficiently handles driver assignments even as the number of drivers scales to 100,000.

#### **b. Asynchronous Assignment**

- **Implementation**: The matching and assignment of drivers to bookings are handled asynchronously to prevent blocking operations.
  
- **Process**:
  - **Goroutine Utilization**: Assignments are processed in separate goroutines, allowing the system to handle multiple assignments concurrently.
  
- **Benefits**:
  - **Throughput**: Enhances the ability to manage numerous booking requests simultaneously.
  - **Responsiveness**: Maintains quick booking confirmations and driver assignments without delay.

### Pricing Model

The **Pricing Service** is a fundamental component of the **On-Demand Logistics Platform for Goods Transportation**, responsible for determining the cost of transporting goods from a user's pickup location to the designated dropoff location. The pricing mechanism ensures that fares are fair, competitive, and dynamically adjusted based on real-time factors such as distance, vehicle type, demand, and supply. This document outlines the methodology and formulas used to calculate transportation costs within the platform.


#### 1. Base Price Calculation

The base price forms the foundational cost of a transportation booking, calculated primarily based on the distance between the pickup and dropoff locations and the type of vehicle selected by the user.

##### a. Distance Measurement

- **Calculation Method:** The distance between the pickup and dropoff points is measured in kilometers (km).
  
- **Measurement Approach:**
  - **Haversine Formula:** A mathematical formula used to calculate the great-circle distance between two points on the Earth's surface, providing an approximation of the shortest distance over the earth’s surface.
  - **Alternative Methods:** Integration with external APIs (e.g., Google Maps Distance Matrix API) for more accurate distance measurements.

##### b. Rate per Kilometer Based on Vehicle Type

Different vehicle types incur varying operational costs and provide different service levels. Consequently, the rate per kilometer varies accordingly.

- **Vehicle Types and Rates:**

  | Vehicle Type | Rate per Kilometer (USD/km) |
  |--------------|-----------------------------|
  | Bike         | $6.00                        |
  | Car          | $12.00                       |
  | Van          | $18.00                       |
  | **Default**  | $30.00                       |

- **Determination Logic:** The rate is selected based on the vehicle type chosen by the user for the booking. If the vehicle type does not match predefined categories (bike, car, van), a default rate is applied.

##### c. Base Price Formula

The base price is calculated using the following formula:

```
Base Price = Distance (km) × Rate per km
```

- **Example Calculation:**
  - **Distance:** 10 km
  - **Vehicle Type:** Car ($12.00/km)

  ```
  Base Price = 10 km × $12.00/km = $120.00
  ```

#### 2. Surge Multiplier Application

To ensure optimal resource utilization and manage high-demand scenarios, a surge multiplier is applied to the base price. This dynamic adjustment reflects the balance between active bookings (demand) and available drivers (supply), as well as time-based factors such as peak hours.

##### a. Demand-Supply Ratio-Based Surge

- **Active Bookings:** The number of ongoing bookings at any given time.
  
- **Available Drivers:** The number of drivers currently available to accept new bookings.

- **Ratio Calculation:**

  ```
  Ratio = Active Bookings / Available Drivers
  ```

- **Multiplier Determination:**

  | Ratio Range                | Surge Multiplier |
  |----------------------------|------------------|
  | Ratio > 1.5                | 1.5              |
  | 1.0 < Ratio ≤ 1.5         | 1.2              |
  | Ratio ≤ 1.0               | 1.0              |

- **Special Condition:**
  
  - **Maximum Surge:** If there are **no available drivers** (Available Drivers = 0), the surge multiplier is set to **2.0** to reflect the high demand and scarcity of drivers.

##### b. Time-Based Surge

Certain times of the day experience higher demand, such as evenings or weekends. To accommodate these fluctuations, an additional surge multiplier is applied during peak hours.

- **Peak Hours Defined:** 6:00 PM to 9:00 PM (18:00 to 21:00)

- **Multiplier Application:**
  
  - **Condition:** If the current time falls within peak hours and the demand-supply ratio does not already impose a higher surge multiplier, an additional surge of **1.3** is applied.

##### c. Surge Multiplier Formula

The final surge multiplier is determined based on the following logic:

1. **Calculate Demand-Supply Ratio:**

   ```
   Ratio = Active Bookings / Available Drivers
   ```

2. **Determine Base Surge Multiplier:**

   ```
   Surge Multiplier =
   ┌ 2.0, if Available Drivers = 0
   ├ 1.5, if Ratio > 1.5
   ├ 1.2, if 1.0 < Ratio ≤ 1.5
   └ 1.0, if Ratio ≤ 1.0
   ```

3. **Apply Time-Based Surge (if applicable):**

   ```
   Final Surge Multiplier =
   ┌ Base Surge Multiplier × 1.3, if current time is within peak hours and Base Surge Multiplier ≤ 1.3
   └ Base Surge Multiplier, otherwise
   ```

- **Final Surge Multiplier Logic:**
  
  - **Example 1:**
    - **Active Bookings:** 30
    - **Available Drivers:** 10
    - **Ratio:** 3.0 (>1.5)
    - **Base Surge Multiplier:** 1.5
    - **Peak Hours:** Yes
    - **Final Surge Multiplier:** max(1.5, 1.3) = 1.5 (No additional surge since base surge is higher)

  - **Example 2:**
    - **Active Bookings:** 12
    - **Available Drivers:** 10
    - **Ratio:** 1.2 (>1.0 but ≤ 1.5)
    - **Base Surge Multiplier:** 1.2
    - **Peak Hours:** Yes
    - **Final Surge Multiplier:** 1.2 × 1.3 = 1.56

#### 3. Final Price Calculation

The final price presented to the user is a combination of the base price and the surge multiplier, reflecting both the fundamental cost of transportation and dynamic factors influencing pricing.

##### Final Price Formula

```
Final Price = Base Price × Surge Multiplier
```

- **Rounding:** The final price is rounded to two decimal places to conform to standard currency formatting.

##### Example Calculation

- **Base Price:** 120.00
- **Surge Multiplier:** 1.5

```
Final Price = 120.00 × 1.5 = 180.00
```

- **Rounded Final Price:** 180.00


### Major Design Decisions and Trade-offs

**1. Modular Monolithic Architecture**

* **Decision**: The system is structured into modular components within a monolithic application. Packages are organized by functionality (e.g., handlers, services, repositories), promoting separation of concerns and maintainability.

* **Trade-off**: While microservices offer scalability benefits, they introduce complexity in deployment and inter-service communication. A modular monolith simplifies development and deployment while still allowing for future refactoring into microservices if needed.

**2. MongoDB as the Database**

* **Decision**: MongoDB was selected for its ability to handle large volumes of data, flexible schema design, and built-in support for sharding and replication.

* **Trade-off**: MongoDB's flexibility comes at the cost of weaker transactional guarantees compared to relational databases, requiring careful handling to maintain data consistency.

**3. Optional Use of NATS Messaging**

* **Decision**: The system supports NATS as an alternative messaging system for distributed communication, configurable via config.yaml.

* **Trade-off**: Incorporating NATS adds complexity but offers scalability benefits for inter-service communication in a distributed environment.

**4. Authentication with JWT**

* **Decision**: JSON Web Tokens (JWT) are used for stateless authentication, eliminating the need for server-side session management and enabling horizontal scaling.

* **Trade-off**: JWT tokens can become invalid if not properly managed (e.g., token revocation is challenging), requiring additional mechanisms for enhanced security.

### **Future Enhancements for Scalability**

#### **a. Caching Strategies**

- **Potential Implementation**: Integrating an in-memory caching layer (e.g., **Redis**) to store frequently accessed data such as pricing factors, vehicle types, and user sessions.
  
- **Benefits**:
  - **Reduced Latency**: Accelerates data retrieval, improving response times for frequently accessed endpoints.
  - **Database Load Reduction**: Decreases the number of read operations on MongoDB, enhancing overall database performance.

#### **b. Microservices Refactoring**

- **Potential Implementation**: As the platform grows, refactoring the monolithic backend into microservices can further enhance scalability and maintainability.
  
- **Benefits**:
  - **Independent Scaling**: Allows individual services to scale based on their specific load and performance requirements.
  - **Improved Fault Isolation**: Limits the impact of failures to individual services, enhancing system resilience.



### Conclusion

The implemented system effectively addresses the requirements for scalability and high-performance real-time data handling. By leveraging Go's concurrency model, efficient use of MongoDB, and real-time communication via WebSockets, the platform is equipped to manage high-volume traffic and provide a seamless experience for users and drivers. The design decisions balance performance, scalability, and maintainability, ensuring that the system can grow and adapt to future demands.

