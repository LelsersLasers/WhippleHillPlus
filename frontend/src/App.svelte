<script>
	import Modal from "./Modal.svelte";

	export let data;
	export let api;

	let assignments = [];
	let classes = [];
	let user = {};
	data
		.then((res) => res.json())
		.then((data) => {
			assignments = data["assignments"];
			classes = data["classes"];
			user = data["user"];
		});

	$: {
		classes = classes.sort((a, b) => a.name.localeCompare(b.name));
	}
	$: {
		function sortAssignments(a, b) {
			// First by date
			if (a.due_date < b.due_date) return -1;
			if (a.due_date > b.due_date) return 1;
			
			// Then by time
			if (a.due_time != "" && b.due_time == "") return -1;
			if (a.due_time == "" && b.due_time != "") return 1;
			if (a.due_time < b.due_time) return -1;
			if (a.due_time > b.due_time) return 1;

			// Then by class
			const class_a = classes.find((c) => c.id === a.class_id);
			const class_b = classes.find((c) => c.id === b.class_id);

			if (class_a.name < class_b.name) return -1;
			if (class_a.name > class_b.name) return 1;

			// Then by status
			const status_weight = {
				"Completed":   2,
				"In Progress": 1,
				"Not Started": 0,
			};
			if (status_weight[a.status] < status_weight[b.status]) return -1;
			if (status_weight[a.status] > status_weight[b.status]) return 1;

			// Then by type
			const type_weight = {
				"Other":    5,
				"Homework": 4,
				"Project":  3,
				"Paper":    2,
				"Quiz":     1,
				"Test":     0,
			};
			if (type_weight[a.type] < type_weight[b.type]) return -1;
			if (type_weight[a.type] > type_weight[b.type]) return 1;

			// lastly by name
			return a.name.localeCompare(b.name);
		}
		assignments = assignments.sort(sortAssignments);
	}

	
	let showCreateClassModal = false;

	let showUpdateClassModal = false;
	let updateClassModalName = "";
	let updateClassModalId = "";

	let showCreateAssignmentModal = false;

	let showUpdateAssignmentModal = false;
	let updateAssignmentModalName = "";
	let updateAssignmentModalDescription = "";
	let updateAssignmentModalAssignedDate = formatDateObj(new Date());
	let updateAssignmentModalDueDate = "";
	let updateAssignmentModalDueTime = "";
	let updateAssignmentModalStatus = "";
	let updateAssignmentModalType = "";
	let updateAssignmentModalClassId = "";
	let updateAssignmentModalId = "";

	let showAssignmentDetailsModal = false;
	let assignmentDetailsModalName = "";
	let assignmentDetailsModalDescription = "";
	let assignmentDetailsModalAssignedDate = "";
	let assignmentDetailsModalDueDate = "";
	let assignmentDetailsModalDueTime = "";
	let assignmentDetailsModalStatus = "";
	let assignmentDetailsModalType = "";
	let assignmentDetailsModalClassId = "";


	function formatDateObj(date) {
		// Date to string in the format "yyyy-MM-dd"
		let year = date.getFullYear();
		let month = String(date.getMonth() + 1).padStart(2, '0');
		let day = String(date.getDate()).padStart(2, '0');
		
		return `${year}-${month}-${day}`;
	}
	
	function formatDateString(str) {
		// "yyyy-MM-dd" -> "m/d/yy"
		let [year, month, day] = str.split("-");
		return `${parseInt(month)}/${parseInt(day)}/${year.slice(2)}`;
	}


	function formDataWithoutReload(e) {
		e.preventDefault();

		const form = e.target;
		const formData = new FormData(form);
		const data = Object.fromEntries(formData);
		
		return data;
	}

	function classFromId(id) {
		return classes.find((c) => c.id === id);
	}

	function deleteClassButton(id) {
		const data = {
			'id': id,
		}
		fetch(`${api}/delete_class`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(data),
		})
			.then((res) => {
				classes = classes.filter((c) => c.id !== id);
				assignments = assignments.filter((a) => a.class_id !== id);
			})
	}
	function updateClassButton(id) {
		const c = classes.find((c) => c.id === id);
		updateClassModalName = c.name;
		updateClassModalId = c.id;
		showUpdateClassModal = true;
	}
	function createClass(e) {
		const data = formDataWithoutReload(e);
		
		fetch(`${api}/create_class`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(data),
		})
			.then((res) =>res.json())
			.then((data) => {
				classes = [...classes, data];
				document.getElementById("createClassModalName").value = "";
				showCreateClassModal = false;
			})
	}
	function updateClass(e) {
		const data = formDataWithoutReload(e);

		fetch(`${api}/update_class`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(data),
		})
			.then((res) => res.json())
			.then((data) => {
				classes = classes.map((c) => {
					if (c.id === data.id) return data;
					else                  return c;
				});
				assignments = assignments;
				updateClassModalName = "";
				updateClassModalId = "";
				showUpdateClassModal = false;
			})
	}

	function deleteAssignmentButton(id) {
		const data = {
			'id': id,
		}
		fetch(`${api}/delete_assignment`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(data),
		})
			.then((res) => {
				assignments = assignments.filter((a) => a.id !== id);
			
				updateAssignmentModalName = "";
				updateAssignmentModalDescription = "";
				updateAssignmentModalAssignedDate = formatDateObj(new Date());
				updateAssignmentModalDueDate = "";
				updateAssignmentModalDueTime = "";
				updateAssignmentModalStatus = "";
				updateAssignmentModalType = "";
				updateAssignmentModalClassId = "";
				showUpdateAssignmentModal = false;
			})
	}
	function updateAssignmentButton(id) {
		const a = assignments.find((a) => a.id === id);
		updateAssignmentModalName = a.name;
		updateAssignmentModalDescription = a.description;
		console.log(a.description);
		updateAssignmentModalAssignedDate = a.assigned_date.slice(0, "yyyy-MM-dd".length);
		updateAssignmentModalDueDate = a.due_date.slice(0, "yyyy-MM-dd".length);
		updateAssignmentModalDueTime = a.due_time;
		updateAssignmentModalStatus = a.status;
		updateAssignmentModalType = a.type;
		updateAssignmentModalClassId = a.class_id;
		updateAssignmentModalId = a.id;
		showUpdateAssignmentModal = true;
	}
	function assignmentDetailsButton(id) {
		const a = assignments.find((a) => a.id === id);
		assignmentDetailsModalName = a.name;
		assignmentDetailsModalDescription = a.description;
		assignmentDetailsModalAssignedDate = a.assigned_date;
		assignmentDetailsModalDueDate = a.due_date;
		assignmentDetailsModalDueTime = a.due_time;
		assignmentDetailsModalStatus = a.status;
		assignmentDetailsModalType = a.type;
		assignmentDetailsModalClassId = a.class_id;
		showAssignmentDetailsModal = true;
	}
	function createAssignment(e) {
		const data = formDataWithoutReload(e);

		fetch(`${api}/create_assignment`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(data),
		})
			.then((res) => res.json())
			.then((data) => {
				assignments = [...assignments, data];
				document.getElementById("createAssignmentModalName").value = "";
				document.getElementById("createAssignmentModalDescription").value = "";
				document.getElementById("createAssignmentModalAssignedDate").valueAsDate = new Date();
				document.getElementById("createAssignmentModalDueDate").value = "";
				document.getElementById("createAssignmentModalDueTime").value = "";
				document.getElementById("createAssignmentModalStatus").value = "Not Started";
				document.getElementById("createAssignmentModalType").value = "Homework";
				// document.getElementById("createAssignmentModalClassId").value = "";
				showCreateAssignmentModal = false;
			})
	};
	function statusAssignment(e, id) {
		const data = {
			'id': id,
			'status': e.target.value,
		}

		fetch(`${api}/status_assignment`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(data),
		})
			.then((res) => res.json())
			.then((data) => {
				assignments = assignments.map((a) => {
					if (a.id === data.id) return data;
					else                  return a;
				});
			})
	}
	function updateAssignment(e) {
		const data = formDataWithoutReload(e);

		fetch(`${api}/update_assignment`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(data),
		})
			.then((res) => res.json())
			.then((data) => {
				assignments = assignments.map((a) => {
					if (a.id === data.id) return data;
					else                  return a;
				});
				updateAssignmentModalName = "";
				updateAssignmentModalDescription = "";
				updateAssignmentModalAssignedDate = formatDateObj(new Date());
				updateAssignmentModalDueDate = "";
				updateAssignmentModalDueTime = "";
				updateAssignmentModalStatus = "";
				updateAssignmentModalType = "";
				updateAssignmentModalClassId = "";
				showUpdateAssignmentModal = false;
			})
	}



	addEventListener("DOMContentLoaded", () => {
		document.getElementById("createClass").addEventListener("submit", createClass);
		document.getElementById("updateClass").addEventListener("submit", updateClass);

		document.getElementById("createAssignmentModalAssignedDate").valueAsDate = new Date();
		document.getElementById("createAssignment").addEventListener("submit", createAssignment);
		document.getElementById("updateAssignment").addEventListener("submit", updateAssignment);
	});
</script>


<style>
.description {
	white-space: pre-line;
}
</style>




<h1>Welcome, {user.name}!</h1>


<h2>Your Classes</h2>
<table>
	<tr>
		<th>ID</th>
		<th>Name</th>
		<th>Edit</th>
		<th>Delete</th>
	</tr>
	{#each classes as c (c.id)}
		<tr>
			<td>{c.id}</td>
			<td>{c.name}</td>
			<td>
				<button type="button" on:click={() => updateClassButton(c.id)}>Edit</button>
			</td>
			<td>
				<button type="button" on:click={() => deleteClassButton(c.id)}>Delete</button>
			</td>
		</tr>
	{/each}
</table>

<h2>Your Assignments</h2>
<table>
	<tr>
		<th>ID</th>
		<th>Class</th>
		<th>Type</th>
		<th>Name</th>
		<th>Assigned</th>
		<th>Due</th>
		<th>Status</th>
		<th>Edit</th>
	</tr>
	{#each assignments as a (a.id)}
		<tr>
			<td>{a.id}</td>
			<td>{classFromId(a.class_id).name}</td>
			<td>{a.type}</td>
			<td>
				<button type="button" on:click={() => assignmentDetailsButton(a.id)}>{a.name}</button>
			</td>
			<td>{formatDateString(a.assigned_date)}</td>
			{#if a.due_time != ""}
				<td>{formatDateString(a.due_date)} - {a.due_time}</td>
			{:else}
				<td>{formatDateString(a.due_date)}</td>
			{/if}

			<td>
				<select value={a.status} on:input={(e) => statusAssignment(e, a.id)}>
					<option value="Not Started">Not Started</option>
					<option value="In Progress">In Progress</option>
					<option value="Completed">Completed</option>
				</select>
			</td>

			<td>
				<button type="button" on:click={() => updateAssignmentButton(a.id)}>Edit</button>
			</td>
		</tr>
	{/each}
</table>



<button type="button" on:click={() => showCreateClassModal = true}>Create Class</button>
<Modal bind:showModal={showCreateClassModal}>
	<h2>Create Class</h2>
	<form id="createClass">
		<label for="name">Name:</label>
		<input type="text" id="createClassModalName" name="name" required>
		<input type="hidden" name="user_id" value={user.id}>
		<button type="submit">Create</button>
	</form>
</Modal>

<Modal bind:showModal={showUpdateClassModal}>
	<h2>Update Class</h2>
	<form id="updateClass">
		<label for="name">Name:</label>
		<input type="text" id="updateClassModalName" name="name" bind:value={updateClassModalName} required>
		<input type="hidden" name="id" bind:value={updateClassModalId}>
		<button type="submit">Update</button>
	</form>
</Modal>



<button type="button" on:click={() => showCreateAssignmentModal = true}>Create Assignment</button>
<Modal bind:showModal={showCreateAssignmentModal}>
	<h2>Create Assignment</h2>
	<form id="createAssignment">
		<label for="name">Name:</label>
		<input type="text" id="createAssignmentModalName" name="name" required>
		<br />
		
		<label for="description">Description:</label>
		<textarea id="createAssignmentModalDescription" name="description"></textarea>
		
		<label for="assigned_date">Assigned Date:</label>
		<input type="date" id="createAssignmentModalAssignedDate" name="assigned_date" required>
		<br />

		<label for="due_date">Due Date:</label>
		<input type="date" id="createAssignmentModalDueDate" name="due_date" required>
		<br />

		<label for="due_time">Due Time:</label>
		<input type="time" id="createAssignmentModalDueTime" name="due_time">
		<br />
		
		<label for="status">Status:</label>
		<select id="createAssignmentModalStatus" name="status">
			<option value="Not Started" selected="selected">Not Started</option>
			<option value="In Progress">In Progress</option>
			<option value="Completed">Completed</option>
		</select>
		<br />

		<label for="type">Type:</label>
		<select id="createAssignmentModalType" name="type">
			<option value="Homework" selected="selected">Homework</option>
			<option value="Quiz">Quiz</option>
			<option value="Test">Test</option>
			<option value="Project">Project</option>
			<option value="Paper">Paper</option>
			<option value="Other">Other</option>
		</select>
		<br />

		<label for="class_id">Class:</label>
		<select id="createAssignmentModalClassId" name="class_id">
			{#each classes as c (c.id)}
				<option value={c.id}>{c.name}</option>
			{/each}
		</select>

		<br />
		<button type="submit">Create</button>
	</form>
</Modal>

<Modal bind:showModal={showUpdateAssignmentModal}>
	<h2>Update Assignment</h2>
	<form id="updateAssignment">
		<label for="name">Name:</label>
		<input type="text" id="updateAssignmentModalName" name="name" bind:value={updateAssignmentModalName} required>
		<br />
		
		<label for="description">Description:</label>
		<textarea id="updateAssignmentModalDescription" name="description" bind:value={updateAssignmentModalDescription}></textarea>
		
		<label for="assigned_date">Assigned Date:</label>
		<input type="date" id="updateAssignmentModalAssignedDate" name="assigned_date" bind:value={updateAssignmentModalAssignedDate} required>
		<br />

		<label for="due_date">Due Date:</label>
		<input type="date" id="updateAssignmentModalDueDate" name="due_date" bind:value={updateAssignmentModalDueDate} required>
		<br />

		<label for="due_time">Due Time:</label>
		<input type="time" id="updateAssignmentModalDueTime" name="due_time" bind:value={updateAssignmentModalDueTime}>
		<br />
		
		<label for="status">Status:</label>
		<select id="updateAssignmentModalStatus" name="status" bind:value={updateAssignmentModalStatus}>
			<option value="Not Started" selected="selected">Not Started</option>
			<option value="In Progress">In Progress</option>
			<option value="Completed">Completed</option>
		</select>
		<br />

		<label for="type">Type:</label>
		<select id="updateAssignmentModalType" name="type" bind:value={updateAssignmentModalType}>
			<option value="Homework" selected="selected">Homework</option>
			<option value="Quiz">Quiz</option>
			<option value="Test">Test</option>
			<option value="Project">Project</option>
			<option value="Paper">Paper</option>
			<option value="Other">Other</option>
		</select>
		<br />

		<label for="class_id">Class:</label>
		<select id="updateAssignmentModalClassId" name="class_id" bind:value={updateAssignmentModalClassId}>
			{#each classes as c (c.id)}
				<option value={c.id}>{c.name}</option>
			{/each}
		</select>

		<input type="hidden" name="id" value={updateAssignmentModalId}>

		<br />
		<button type="submit">Update</button>
		<button type="button" on:click={() => deleteAssignmentButton(updateAssignmentModalId)}>Delete</button>
	</form>
</Modal>

<Modal bind:showModal={showAssignmentDetailsModal}>
	<h2>Assignment Details</h2>
	<p>Name: {assignmentDetailsModalName}</p>
	<p class="description">Description: <br/>{assignmentDetailsModalDescription}</p>
	<p>Assigned Date: {assignmentDetailsModalAssignedDate}</p>
	<p>Due Date: {assignmentDetailsModalDueDate} - {assignmentDetailsModalDueTime}</p>
	<p>Status: {assignmentDetailsModalStatus}</p>
	<p>Type: {assignmentDetailsModalType}</p>
	<p>Class: {classFromId(assignmentDetailsModalClassId)?.name}</p>
</Modal>