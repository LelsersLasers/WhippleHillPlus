<script>
	export let data;

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
		
</script>

<style>
</style>

<main>
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



</main>

