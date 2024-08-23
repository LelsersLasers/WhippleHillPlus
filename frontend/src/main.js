import App from './App.svelte';

const app = new App({
	target: document.body,
	props: {
		data: fetch("http://localhost:8080/home_data")
	}
});

export default app;