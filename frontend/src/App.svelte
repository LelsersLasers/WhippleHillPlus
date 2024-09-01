<script>
	import Modal from "./Modal.svelte";
	import { fly } from "svelte/transition"; 

	export let data;
	export let api;

	let page = "assignments";

	let assignments = [];
	let classes = [];
	let user = {};

	let shownAssignments = [];
	let classFilter = [];

	let unique = {};

	data
		.then((res) => res.json())
		.then((data) => {
			assignments = data["assignments"];

			classes = data["classes"];
			classFilter = classes.map((c) => c.id);
			user = data["user"];

			unique = {};
			assignments.forEach((a) => {
				unique[a.id] = 0;
			});
		});

	function sortClasses(a, b) {
		// Sort "other" to the bottom
		// Otherwise alphabetically
		if (a.name === b.name) return 0;
		if (a.name.toLowerCase() === "other") return 1;
		if (b.name.toLowerCase() === "other") return -1;
		return a.name.localeCompare(b.name);
	}

	$: {
		classes = classes.sort(sortClasses);
	}
	$: {
		classFilter = classFilter.sort((a, b) => classFromId(a).name.localeCompare(classFromId(b).name));
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

			// Then by status
			const statusSeight = {
				"Completed":   1,
				"In Progress": 0,
				"Not Started": 0,
			};
			if (statusSeight[a.status] < statusSeight[b.status]) return -1;
			if (statusSeight[a.status] > statusSeight[b.status]) return 1;

			// Then by type
			const typeWeight = {
				"Other":    5,
				"Homework": 4,
				"Project":  3,
				"Paper":    2,
				"Quiz":     1,
				"Test":     0,
			};
			if (typeWeight[a.type] < typeWeight[b.type]) return -1;
			if (typeWeight[a.type] > typeWeight[b.type]) return 1;

			// Then by class
			const classA = classes.find((c) => c.id === a.class_id);
			const classB = classes.find((c) => c.id === b.class_id);
			const classSort = sortClasses(classA, classB);
			if (classSort !== 0) return classSort;

			// lastly by name
			return a.name.localeCompare(b.name);
		}
		function convertToLocalTime(a) {
			if (a.due_date.endsWith("Z")) {
				a.due_date = a.due_date.slice(0, -1);
			}
			if (a.assigned_date.endsWith("Z")) {
				a.assigned_date = a.assigned_date.slice(0, -1);
			}
		}
		assignments.forEach(convertToLocalTime);
		assignments.sort(sortAssignments);
		assignments = assignments;
	}

	function localDate() {
		const date = new Date();
		const timeOffset = date.getTimezoneOffset();
		const hoursOffset = Math.floor(timeOffset / 60);
		const minutesOffset = timeOffset % 60;
		date.setHours(date.getHours() - hoursOffset);
		date.setMinutes(date.getMinutes() - minutesOffset);
		return date;
	}

	
	let showCreateClassModal = false;
	let createClassModalButton = true;

	let showUpdateClassModal = false;
	let updateClassModalName = "";
	let updateClassModalId = "";
	let updateClassModalButton = true;

	let showDeleteClassModal = false;
	let deleteClassModalName = "";
	let deleteClassModalId = "";
	let deleteClassModalTimer = 15;
	let deleteClassModalTimerInterval = null;
	let deleteClassModalButton = true;


	let showCreateAssignmentModal = false;
	let createAssignmentModalButton = true;

	let showUpdateAssignmentModal = false;
	let updateAssignmentModalName = "";
	let updateAssignmentModalDescription = "";
	let updateAssignmentModalAssignedDate = formatDateObj(localDate());
	let updateAssignmentModalDueDate = "";
	let updateAssignmentModalDueTime = "";
	let updateAssignmentModalStatus = "";
	let updateAssignmentModalType = "";
	let updateAssignmentModalClassId = "";
	let updateAssignmentModalId = "";
	let updateAssignmentModalButton = true;
	let updateAssignmentModalDeleteButton = true;

	let showAssignmentDetailsModal = false;
	let assignmentDetailsModalName = "";
	let assignmentDetailsModalId = "";
	let assignmentDetailsModalDescription = "";
	let assignmentDetailsModalAssignedDate = "";
	let assignmentDetailsModalDueDate = "";
	let assignmentDetailsModalDueTime = "";
	let assignmentDetailsModalStatus = "";
	let assignmentDetailsModalType = "";
	let assignmentDetailsModalClassId = "";

	let showDateFilterModal = false;
	let dateFilter = "future";

	let showClassFilterModal = false;

	let showStatusFilter = false;
	let statusFilter = ["Not Started", "In Progress", "Completed"];

	const today = new Date();
	today.setHours(0, 0, 0, 0);

	const pastSunday = new Date(today);
	pastSunday.setDate(today.getDate() - today.getDay());
	
	const nextSunday = new Date(today);
	if (7 - today.getDay() < 2) nextSunday.setDate(today.getDate() + (14 - today.getDay()));
	else                        nextSunday.setDate(today.getDate() + (7  - today.getDay()));
	
	let dateWeekStart = formatDateObj(pastSunday);
	let dateWeekEnd = formatDateObj(nextSunday);

	let dateStart = formatDateObj(pastSunday);
	let dateEnd = formatDateObj(nextSunday);

	function missingCheck(a) {
		// due date is in the past and status is not completed
		if (a.status == "Completed") return false;

		const dueDate = new Date(a.due_date);
		if (a.due_time != "") {
			const todayWithTime = new Date();

			const dueTime = a.due_time.split(":");
			dueDate.setHours(dueTime[0]);
			dueDate.setMinutes(dueTime[1]);

			return dueDate < todayWithTime;
		} else {
			return dueDate < today;
		}
	}

	function assignmentToColor(a) {
		const missing = missingCheck(a);
		if (missing) return "#bf616a";
		const status_color = {
			"Not Started": "#b48ead",
			"In Progress": "#ebcb8b",
			"Completed": "#a3be8c",
		};
		return status_color[a.status];
	}

	$: {
		function classFilterCheck(a) {
			return classFilter.includes(a.class_id);
		}
		function statusFilterCheck(a) {
			return statusFilter.includes(a.status);
		}

		if (dateFilter == "all") {
			shownAssignments = assignments;
		} else if (dateFilter == "week") {
			shownAssignments = assignments.filter((a) => {
				const dueDate = new Date(a.due_date);
				const assignedDate = new Date(a.assigned_date);
				const dueDateInRange = dueDate >= new Date(dateWeekStart) && dueDate <= new Date(dateWeekEnd);
				const assignedDateInRange = assignedDate >= new Date(dateWeekStart) && assignedDate <= new Date(dateWeekEnd);
				return (classFilterCheck(a) && statusFilterCheck(a) && (dueDateInRange || assignedDateInRange)) || missingCheck(a);
			});
		} else if (dateFilter == "future") {
			shownAssignments = assignments.filter((a) => {
				const dueDate = new Date(a.due_date);
				return (classFilterCheck(a) && statusFilterCheck(a) && dueDate >= today) || missingCheck(a);
			});
		} else if (dateFilter == "range") {
			shownAssignments = assignments.filter((a) => {
				const dueDate = new Date(a.due_date);
				const assignedDate = new Date(a.assigned_date);
				const dueDateInRange = dueDate >= new Date(dateStart) && dueDate <= new Date(dateEnd);
				const assignedDateInRange = assignedDate >= new Date(dateStart) && assignedDate <= new Date(dateEnd);
				return (classFilterCheck(a) && statusFilterCheck(a) && (dueDateInRange || assignedDateInRange)) || missingCheck(a);
			});
		}
	}


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
		deleteClassModalButton = false;

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
				classFilter = classFilter.filter((c) => c !== id);

				deleteClassModalName = "";
				deleteClassModalId = "";
				deleteClassModalTimer = 0;
				showDeleteClassModal = false;
				deleteClassModalButton = true;
			})
	}
	function deleteModalButton(id) {
		const c = classes.find((c) => c.id === id);
		deleteClassModalName = c.name;
		deleteClassModalId = c.id;
		deleteClassModalTimer = 15;
		if (deleteClassModalTimerInterval) clearInterval(deleteClassModalTimerInterval);
		deleteClassModalTimerInterval = setInterval(() => {
			deleteClassModalTimer -= 1;
			if (deleteClassModalTimer <= 0) {
				clearInterval(deleteClassModalTimerInterval);
			}
		}, 1000);
		showDeleteClassModal = true;
	}
	function updateClassButton(id) {
		const c = classes.find((c) => c.id === id);
		updateClassModalName = c.name;
		updateClassModalId = c.id;
		showUpdateClassModal = true;
	}
	function createClass(e) {
		const data = formDataWithoutReload(e);
		createClassModalButton = false;
		
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
				classFilter = [...classFilter, data.id];
				document.getElementById("createClassModalName").value = "";
				showCreateClassModal = false;
				createClassModalButton = true;
			})
	}
	function updateClass(e) {
		const data = formDataWithoutReload(e);
		updateClassModalButton = false;

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
				updateClassModalButton = true;
			})
	}

	function deleteAssignmentButton(id) {
		const data = {
			'id': id,
		}
		updateAssignmentModalDeleteButton = false;

		fetch(`${api}/delete_assignment`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(data),
		})
			.then((res) => {
				assignments = assignments.filter((a) => a.id !== id);
				unique[id] += 1;
			
				updateAssignmentModalName = "";
				updateAssignmentModalDescription = "";
				updateAssignmentModalAssignedDate = formatDateObj(localDate());
				updateAssignmentModalDueDate = "";
				updateAssignmentModalDueTime = "";
				updateAssignmentModalStatus = "";
				updateAssignmentModalType = "";
				updateAssignmentModalClassId = "";
				showUpdateAssignmentModal = false;
				updateAssignmentModalDeleteButton = true;
			})
	}
	function updateAssignmentButton(id) {
		const a = assignments.find((a) => a.id === id);
		updateAssignmentModalName = a.name;
		updateAssignmentModalDescription = a.description;
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
		assignmentDetailsModalId = a.id;
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
		createAssignmentModalButton = false;

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
				unique[data.id] += 1;

				document.getElementById("createAssignmentModalName").value = "";
				document.getElementById("createAssignmentModalDescription").value = "";
				document.getElementById("createAssignmentModalAssignedDate").valueAsDate = localDate();
				document.getElementById("createAssignmentModalDueDate").value = "";
				document.getElementById("createAssignmentModalDueTime").value = "";
				document.getElementById("createAssignmentModalStatus").value = "Not Started";
				document.getElementById("createAssignmentModalType").value = "Homework";
				// document.getElementById("createAssignmentModalClassId").value = "";
				showCreateAssignmentModal = false;
				createAssignmentModalButton = true;
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
				unique[data.id] += 1;
			})
	}
	function updateAssignment(e) {
		const data = formDataWithoutReload(e);
		updateAssignmentModalButton = false;

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
				unique[data.id] += 1;

				updateAssignmentModalName = "";
				updateAssignmentModalDescription = "";
				updateAssignmentModalAssignedDate = formatDateObj(localDate());
				updateAssignmentModalDueDate = "";
				updateAssignmentModalDueTime = "";
				updateAssignmentModalStatus = "";
				updateAssignmentModalType = "";
				updateAssignmentModalClassId = "";
				showUpdateAssignmentModal = false;
				updateAssignmentModalButton = true;

				assignmentDetailsModalName = data.name;
				assignmentDetailsModalDescription = data.description;
				assignmentDetailsModalAssignedDate = data.assigned_date;
				assignmentDetailsModalDueDate = data.due_date;
				assignmentDetailsModalDueTime = data.due_time;
				assignmentDetailsModalStatus = data.status;
				assignmentDetailsModalType = data.type;
				assignmentDetailsModalClassId = data.class_id;
			})
	}



	addEventListener("DOMContentLoaded", () => {
		document.getElementById("createClass").addEventListener("submit", createClass);
		document.getElementById("updateClass").addEventListener("submit", updateClass);

		document.getElementById("createAssignmentModalAssignedDate").valueAsDate = localDate();
		document.getElementById("createAssignment").addEventListener("submit", createAssignment);
		document.getElementById("updateAssignment").addEventListener("submit", updateAssignment);
	});
</script>


<style>

#holder {
	width: 100%;
	margin: 0 auto;
}

@media (min-width: 800px) {
	#holder {
		width: 80%;
	}
}


.description {
	white-space: pre-line;
}

table {
	/* width: 80%;
	margin: 0 auto; */
	width: 100%;
	border-collapse: collapse;
}

th {
	background-color: #afafaf;
	text-align: left;
	padding-left: 0;
	border-bottom: 2px black solid;
}

tr:nth-child(even) {
	background-color: #f4f4f4;
}

.breakWord {
	word-break: break-word;
}


.zeroWidth {
	width: 0;
}
.classWidth {
	width: 15vw;
}

.untightPadding {
	padding-right: 15px;
}

.pointer {
	cursor: pointer;
}


input[type="text"] {
	width: 100%;
}
textarea {
	width: 100%;
	height: 5em;
}


select:active {
	background-color: revert !important;
}

button:disabled {
	animation: bob 0.5s infinite ease-in-out;
}

@keyframes bob {
	0% {
		transform: translateY(0);
	}
	50% {
		transform: translateY(-5px);
	}
	100% {
		transform: translateY(0);
	}
}

</style>



<div id="holder">
<h1>Welcome, {user.name}!</h1>
<button type="button" on:click={() => window.location.href = "/logout_user"}>Logout</button>

{#if page == "classes"}
	<h2>Your Classes</h2>
	<button type="button" on:click={() => page = "assignments"}>View Assignments</button>
	<button type="button" on:click={() => showCreateClassModal = true}>Create Class</button>

	<table>
		<tr>
			<th>Name</th>
			<th class="zeroWidth"></th>
			<th class="zeroWidth"></th>
		</tr>
		{#each classes as c (c.id)}
			<tr
				in:fly|global={{ duration: 300, x: -200 }}
			>
				<td class="breakWord">{c.name}</td>
				<td class="zeroWidth">
					<button type="button" on:click={() => updateClassButton(c.id)}>Edit</button>
				</td>
				<td class="zeroWidth">
					<button type="button" on:click={() => deleteModalButton(c.id)}>Delete</button>
				</td>
			</tr>
		{/each}
	</table>
{:else if page == "assignments"}
	<h2>Your Assignments</h2>
	<button type="button" on:click={() => page = "classes"}>View Classes</button>
	<button type="button" on:click={() => showCreateAssignmentModal = true}>Create Assignment</button>
	<button type="button" on:click={() => showDateFilterModal = true}>Date Filter</button>
	<button type="button" on:click={() => showClassFilterModal = true}>Class Filter</button>
	<button type="button" on:click={() => showStatusFilter = true}>Status Filter</button>

	<table>
		<tr>
			<th class="classWidth">Class</th>
			<th class="zeroWidth untightPadding">Type</th>
			<th>Name</th>
			<th class="zeroWidth untightPadding">Assigned</th>
			<th class="zeroWidth">Due</th>
			<th class="zeroWidth"></th>
			<th class="zeroWidth"></th>
		</tr>
		{#each shownAssignments as a (a.id)}
			{#key unique[a.id]}
				<tr
					in:fly|global={{ duration: 300, x: -200 }}
				>
					<td class="breakWord padding">{classFromId(a.class_id).name}</td>
					<td class="untightPadding">{a.type}</td>
					<td>
						<a class="breakWord pointer" on:click={() => assignmentDetailsButton(a.id)}>{a.name}</a>
					</td>
					<td class="untightPadding">{formatDateString(a.assigned_date)}</td>
					{#if a.due_time != ""}
						<td class="untightPadding">{formatDateString(a.due_date)} <br /> {a.due_time}</td>
					{:else}
						<td class="untightPadding">{formatDateString(a.due_date)}</td>
					{/if}

					<td class="zeroWidth untightPadding">
						<select value={a.status} on:input={(e) => statusAssignment(e, a.id)} style="background-color: {assignmentToColor(a)}">
							<option value="Not Started">Not Started</option>
							<option value="In Progress">In Progress</option>
							<option value="Completed">Completed</option>
						</select>
					</td>

					<td class="zeroWidth">
						<button type="button" on:click={() => updateAssignmentButton(a.id)}>Edit</button>
					</td>
				</tr>
			{/key}
		{/each}
	</table>

	<br />
	<br />
	<br />
{/if}

</div>



<Modal bind:showModal={showCreateClassModal}>
	<h2>Create Class</h2>
	<form id="createClass">
		<label for="name">Name:</label>
		<input type="text" id="createClassModalName" name="name" required>
		<input type="hidden" name="user_id" value={user.id}>
		<button type="submit" disabled={!createClassModalButton}>Create</button>
	</form>
</Modal>

<Modal bind:showModal={showUpdateClassModal}>
	<h2>Update Class</h2>
	<form id="updateClass">
		<label for="name">Name:</label>
		<input type="text" id="updateClassModalName" name="name" bind:value={updateClassModalName} required>
		<input type="hidden" name="id" bind:value={updateClassModalId}>
		<button type="submit" disabled={!updateClassModalButton}>Update</button>
	</form>
</Modal>

<Modal bind:showModal={showDeleteClassModal}>
	<h2>Delete Class</h2>
	<p>Are you sure you want to delete the class "{deleteClassModalName}"?</p>
	<p>Deleting the class will delete all assignments associated with it.</p>
	{#if deleteClassModalTimer > 0}
		<p>Wait {deleteClassModalTimer} seconds before deleting.</p>
	{:else}
		<button type="button" on:click={() => deleteClassButton(deleteClassModalId)} disabled={!deleteClassModalButton}>Yes</button>
	{/if}
</Modal>


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
		<button type="submit" disabled={!createAssignmentModalButton}>Create</button>
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
		<button type="submit" disabled={!updateAssignmentModalButton}>Update</button>
		<button type="button" on:click={() => deleteAssignmentButton(updateAssignmentModalId)} disabled={!updateAssignmentModalDeleteButton}>Delete</button>
	</form>
</Modal>

<Modal bind:showModal={showAssignmentDetailsModal}>
	<h2>Assignment Details</h2>
	<p>Name: {assignmentDetailsModalName}</p>
	<p class="description">Description: <br/>{assignmentDetailsModalDescription}</p>
	<p>Assigned Date: {formatDateString(assignmentDetailsModalAssignedDate)}</p>

	{#if assignmentDetailsModalDueTime != ""}
		<p>Due Date: {formatDateString(assignmentDetailsModalDueDate)} - {assignmentDetailsModalDueTime}</p>
	{:else}
		<p>Due Date: {formatDateString(assignmentDetailsModalDueDate)}</p>
	{/if}
	<p>Status: {assignmentDetailsModalStatus}</p>
	<p>Type: {assignmentDetailsModalType}</p>
	<p>Class: {classFromId(assignmentDetailsModalClassId)?.name}</p>
	<button type="button" on:click={() => updateAssignmentButton(assignmentDetailsModalId)}>Edit</button>	
</Modal>


<Modal bind:showModal={showDateFilterModal}>
	<h2>Date Filter</h2>

	<label for="all">
		<input type="radio" id="all" value="all" bind:group={dateFilter}>
		All
	</label>

	<label for="week">
		<input type="radio" id="week" value="week" bind:group={dateFilter}>
		This Week
	</label>

	<label for="future">
		<input type="radio" id="future" value="future" bind:group={dateFilter}>
		Future
	</label>

	<label for="range">
		<input type="radio" id="range" value="range" bind:group={dateFilter}>
		Custom Range
	</label>

	{#if dateFilter == "range"}
		<label for="dateStart">Start Date:</label>
		<input type="date" id="dateStart" bind:value={dateStart}>
		<label for="dateEnd">End Date:</label>
		<input type="date" id="dateEnd" bind:value={dateEnd}>
	{/if}
</Modal>

<Modal bind:showModal={showClassFilterModal}>
	<h2>Class Filter</h2>
	{#each classes as c (c.id)}
		<label for={c.id}>
			<input type="checkbox" id={c.id} value={c.id} bind:group={classFilter}>
			{c.name}
		</label>
	{/each}
</Modal>

<Modal bind:showModal={showStatusFilter}>
	<h2>Status Filter</h2>
	{#each ["Not Started", "In Progress", "Completed"] as s}
		<label for={s}>
			<input type="checkbox" id={s} value={s} bind:group={statusFilter}>
			{s}
		</label>
	{/each}
</Modal>