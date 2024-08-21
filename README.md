# Whipple Hill Clone

A mimic of Bishop's Whipple Hill for assignment and due date management


## Database

- Class
	- id (int, pk)
	- name (string)
- Assignment
	- id (int, pk)
	- name (string)
	- description (string)
	- due_date (date)
	- due_time (time)
	- assigned_date (date)
	- class_id (int, fk)