# Whipple Hill+

A mimic of Bishop's Whipple Hill for assignment and due date management

## Todo

- favicon
- Actually CSS

## Database

- users
	- id (int, pk)
	- username (text)
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
	- POST /create_assignment - Add a new assignment
	- POST /update_assignment - Edit a single assignment
	- POST /delete_assignment - Delete a single assignment
	- POST /status_assignment - Change the status of an assignment
- Class stuff
	- POST /create_class - Add a new class
	- POST /update_class - Edit a single class
	- POST /delete_class - Delete a single class

### Pages

- GET /login - Login form
- GET /register - Signup form
- GET / - Home page
	- Redirect to login if not logged in
	- Svelte