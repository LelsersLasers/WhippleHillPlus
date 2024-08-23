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

	
	let showCreateClassModal = false;

	let showUpdateClassModal = false;
	let updateClassModalName = "";
	let updateClassModalId = "";

	let showCreateAssignmentModal = false;

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
				updateClassModalName = "";
				updateClassModalId = "";
				showUpdateClassModal = false;
			})
		}

	addEventListener("DOMContentLoaded", () => {
		document.getElementById("createClass").addEventListener("submit", createClass);
		document.getElementById("updateClass").addEventListener("submit", updateClass);
	});
</script>


<style>
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
	</tr>
	{#each assignments as a (a.id)}
		<tr>
			<td>{a.id}</td>
			<td>{classFromId(a.class_id).name}</td>
			<td>{a.type}</td>
			<td>{a.name}</td>
			<td>{a.assigned_date}</td>
			<td>{a.due_date} - {a.due_time}</td>
			<td>{a.status}</td>
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

