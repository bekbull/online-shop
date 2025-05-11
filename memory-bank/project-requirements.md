# Project Requirements: Go Microservices Online Shop

## 1. Core System Requirements

### 1.1. Microservices Architecture
* **Language:** Go programming language.
* **Minimum Services:** The system must consist of at least three distinct microservices.
* **Design Principles:** Each microservice should serve a specific domain or functionality, ensuring clear separation of concerns and a modular design.
* **Scalability & Flexibility:** The architecture should be designed with scalability and flexibility in mind.

### 1.2. Communication Between Microservices
* **Technology:** gRPC must be utilized for communication between microservices.
* **Data Exchange:** Protocol buffer messages must be defined to establish language-agnostic communication contracts.
* **Efficiency:** Implement bidirectional communication where appropriate for efficient data exchange.

### 1.3. Public HTTP Endpoints
* **Client Interaction:** Each microservice must expose public HTTP endpoints to interact with external clients (e.g., web frontend, mobile app).
* **Design:** Follow RESTful principles for designing these endpoints, ensuring clarity, consistency, and standard HTTP methods.
* **Functionality:** Endpoints should cover CRUD (Create, Read, Update, Delete) operations and any other relevant functionalities specific to the microservice's domain.

### 1.4. Data Persistence
* **Database Integration:** Integrate a database of choice (e.g., PostgreSQL, MySQL, MongoDB) to persist data for each microservice as needed.
* **Schema Design:** Design database schemas to efficiently store and retrieve data relevant to each microservice's functionalities.
* **Database Operations:** Implement CRUD operations within the microservices to interact with their respective databases.

### 1.5. Concurrency and Error Handling
* **Concurrency:** Leverage Goroutines and channels for implementing concurrency where applicable to enhance performance and responsiveness.
* **Robustness:** Ensure proper error handling throughout the codebase to maintain system robustness and reliability.
* **Error Reporting:** Utilize custom error types (where beneficial) and structured logging to provide informative error messages and aid in debugging.

## 2. Testing and Benchmarking Requirements

### 2.1. Unit Testing
* **Comprehensiveness:** Write comprehensive unit tests for each microservice to validate its functionality.
* **Tooling:** Use the standard Go `testing` package or any preferred Go testing framework.
* **Coverage:** Aim for high test coverage across all major functionalities.

### 2.2. Benchmarking
* **Performance Measurement:** Conduct benchmarking to measure the performance of critical components and API endpoints.
* **Optimization:** Use benchmark results to identify areas for potential optimization.

## 3. Documentation Requirements

### 3.1. System Overview
* Provide clear and concise documentation for the project.
* Include an overview of the system architecture, detailing the roles, responsibilities, and interactions of each microservice.

### 3.2. API Documentation
* Document all public API endpoints.
* Specify request and response formats (e.g., using JSON schemas).
* Detail parameters, headers, and expected status codes.
* **(Bonus)** Use tools like Swagger/OpenAPI to generate interactive API documentation.

### 3.3. Setup and Deployment
* Provide instructions for setting up the project locally, including all dependencies and environment configurations.
* Include clear steps for building and running the project.

## 4. Additional Project Guidelines

### 4.1. Collaboration (If applicable for team projects)
* Assign clear roles and responsibilities within the team.
* Maintain regular communication regarding progress, issues, and challenges.

### 4.2. Solution Quality
* Aim for a well-structured, maintainable, and scalable solution.
* Adhere to Go coding best practices, including code readability, modularity, and performance optimization.

### 4.3. Creativity and Innovation
* Demonstrate creativity and innovation in the project implementation.
* Consider features or enhancements beyond the basic requirements that add value or showcase problem-solving skills.

## 5. Grading Rubrics (For Reference)

The project will be evaluated based on the following criteria:

### 5.1. Architecture (25%)
* **Excellent (5):** Clear separation of concerns, modular design, and effective communication between microservices demonstrated. Architecture demonstrates scalability and flexibility.
* **Good (4):** Microservices architecture is well-defined, with moderate separation of concerns and communication between components. Some improvements could be made for scalability and modularity.
* **Satisfactory (3):** Basic microservices architecture is implemented, but lacks clear separation of concerns or modularity. Communication between microservices may be limited or not fully optimized.
* **Needs Improvement (2):** Microservices architecture is rudimentary, with little consideration for separation of concerns or communication patterns. Significant improvements needed for scalability and flexibility.
* **Poor (1):** Microservices architecture is poorly designed, with little to no separation of concerns or modularity. Communication between microservices is unclear or non-existent.

### 5.2. Functionality (25%)
* **Excellent (5):** All required functionalities are implemented with precision and accuracy. Microservices handle requests efficiently and respond appropriately. Error handling is robust and well-integrated.
* **Good (4):** Most required functionalities are implemented, but with minor issues or gaps in functionality. Microservices generally handle requests and responses effectively, but may have occasional errors or inconsistencies.
* **Satisfactory (3):** Basic functionalities are implemented, but with notable gaps or inconsistencies. Microservices may struggle to handle certain requests or exhibit errors under certain conditions.
* **Needs Improvement (2):** Limited functionalities are implemented, with significant gaps or errors in functionality. Microservices may fail to handle certain requests or exhibit frequent errors.
* **Poor (1):** Essential functionalities are missing or severely flawed. Microservices consistently fail to handle requests or exhibit critical errors.

### 5.3. Testing and Reliability (20%)
* **Excellent (5):** Comprehensive unit tests cover all major functionalities, with high test coverage and effective error handling. Benchmarking results demonstrate performance optimizations.
* **Good (4):** Unit tests cover most major functionalities, with moderate test coverage and adequate error handling. Some benchmarking conducted to evaluate performance.
* **Satisfactory (3):** Basic unit tests are implemented, but with limited coverage or effectiveness. Error handling may be lacking or inconsistent. Limited benchmarking conducted.
* **Needs Improvement (2):** Few unit tests are implemented, with minimal coverage and ineffective error handling. Benchmarking is lacking or not conducted appropriately.
* **Poor (1):** No or very minimal testing is implemented. Errors are rampant, and reliability is severely compromised.

### 5.4. Deployment (15%)
* **Excellent (5):** Microservices are successfully deployed on a cloud platform or container orchestration with demonstrated scalability and fault tolerance.
* **Good (4):** Microservices are deployed, but with minor issues or limitations in scalability or fault tolerance.
* **Satisfactory (3):** Basic deployment is achieved, but with notable gaps in scalability or fault tolerance.
* **Needs Improvement (2):** Deployment is attempted, but with significant issues or failures in scalability or fault tolerance.
* **Poor (1):** Deployment is not achieved or severely flawed.

### 5.5. Creativity and Innovation (15%)
* **Excellent (5):** Project demonstrates innovative features or enhancements beyond basic requirements, showcasing creativity and problem-solving skills.
* **Good (4):** Some additional features or enhancements are implemented, showcasing creativity and initiative.
* **Satisfactory (3):** Minimal creativity or innovation demonstrated. Project meets basic requirements without significant enhancements.
* **Needs Improvement (2):** Little to no creativity or innovation demonstrated. Project lacks additional features or enhancements beyond basic requirements.
* **Poor (1):** No creativity or innovation demonstrated. Project strictly adheres to basic requirements without any additional features or enhancements.