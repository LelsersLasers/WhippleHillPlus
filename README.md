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

### Backend

- User stuff:
	- POST /login_user - Login user
	- POST /register_user - Signup user
	- POST /logout_user - Logout user
- Assignment stuff
	<!-- - GET /assignments - List of all assignments -->
	- POST /create_assignment - Add a new assignment
	- GET /get_assignment?id- View a single assignment (in popup)
	- POST /update_assignment [id] - Edit a single assignment
	- POST /delete_assignment [id] - Delete a single assignment
- Class stuff
	<!-- - GET /classes - List of all classes -->
	- POST /create_class - Add a new class
	- POST /update_class [id] - Edit a single class
	- POST /delete_class [id] - Delete a single class

### Pages

- GET /login - Login form
- GET /register - Signup form
- GET / - Home page
	- Redirect to login if not logged in