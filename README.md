# Whipple Hill Clone

A mimic of Bishop's Whipple Hill for assignment and due date management


## Database

- users
	- id (int, pk)
	- email (text)
	- display_name (text)
	- password_hash (text)
- classes
	- id (int, pk)
	- name (text)
	- user_id (int, fk)
- assignments
	- id (int, pk)
	- name (text)
	- description (text)
	- due_date (date)
	- due_time (time)
	- assigned_date (date)
	- class_id (int, fk)
	- status (text) ["Not Started", "In Progress", "Completed"]
	- type (text) ["Homework", "Quiz", "Test", "Project", "Paper", "Other"]

## Routes

### API

- POST /login - Login user
- POST /register - Signup user
- POST /logout - Logout user
- GET /assignments - List of all assignments
- POST /assignments - Add a new assignment
- GET /assignments/:id - View a single assignment
- PUT /assignments/:id - Edit a single assignment
- DELETE /assignments/:id - Delete a single assignment
- GET /classes - List of all classes
- POST /classes - Add a new class
- PUT /classes/:id - Edit a single class
- DELETE /classes/:id - Delete a single class

### Web

- GET /login - Login form
- GET /register - Signup form
- GET / - Home page
	- Redirect to login if not logged in
- GET /classes/:id - View assignments for a single class