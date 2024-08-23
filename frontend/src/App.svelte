<script>
	export let data;
	export let api;

	let assignments = [];
	let classes = [];
	let user = {};
	data
		.then((res) => res.json())
		.then((data) => {
			console.log(data);
			assignments = data["assignments"];
			classes = data["classes"];
			user = data["user"];
		});

	function classFromId(id) {
		return classes.find((c) => c.id === id);
	}

	addEventListener("DOMContentLoaded", () => {
		document.getElementById("createClass").addEventListener("submit", (e) => {
			e.preventDefault();
			const form = e.target;
			const formData = new FormData(form);
			const data = Object.fromEntries(formData);
			console.log(data);
			fetch(`${api}/create_class`, {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(data),
				// body: formData,
			})
				.then((res) => {
					console.log(res);
					return res.json();
				})
				.then((data) => {
					console.log(data);
					classes = [...classes, data];
				})
				// .catch((err) => console.error(err));
		});
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
	</tr>
	{#each classes as c}
		<tr>
			<td>{c.id}</td>
			<td>{c.name}</td>
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
	{#each assignments as a}
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


<h2>Create Class</h2>
<form id="createClass">
	<label for="name">Name:</label>
	<input type="text" id="name" name="name" required>
	<input type="hidden" name="user_id" value={user.id}>
	<button type="submit">Create</button>
</form>

