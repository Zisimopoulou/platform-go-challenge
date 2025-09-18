A RESTful API for managing user favorites including charts, insights, and audience segments.

## Features
* User Authentication: JWT-based authentication
* CRUD Operations: Create, read, update, and delete favorites
* Support for charts, insights and audiences
* Input validation
* In-Memory Storage

## API Endpoints
Favorites management using bearer token for authentication
* POST	/users/{user}/favorites Create a new favorite	
* GET	/users/{user}/favorites	List all favorites	 
* PUT	/users/{user}/favorites/{id}	Update a favorite 
* DELETE	/users/{user}/favorites/{id}	Delete a favorite

* Login /auth/login

## Asset Types

* Chart
  
{
  "type": "chart",
  "description": "Monthly revenue growth",
  "payload": {
    "title": "Revenue Chart",
    "xAxis": "Months",
    "yAxis": "Revenue ($)",
    "data": [1000, 1500, 2000, 2500]
  }
}

* Insight
  
{
  "type": "insight",
  "description": "Key market insight",
  "payload": {
    "text": "40% of millennials prefer mobile shopping"
  }
}

* Audience
  
{
  "type": "audience",
  "description": "Target demographic",
  "payload": {
    "gender": "Male",
    "birthCountry": "Greece",
    "ageGroup": "24-35",
    "hoursDaily": "3+",
    "purchasesLastMonth": "2"
  }
}


## Instructions

* Add dependencies
  
go get github.com/go-playground/validator/v10

go get github.com/golang-jwt/jwt/v5

go get github.com/davecgh/go-spew/spew@latest

go get github.com/go-chi/chi/v5

go get github.com/go-chi/cors

go mod tidy

* Set up environment variables
  
$env:JWT_SECRET="dev-super-secure-random-secret-32-chars-long!"
