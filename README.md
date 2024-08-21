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