import App from './App.svelte';

const API = "http://64.98.192.13:3003";
// const API = "http://localhost:3003";
const app = new App({
	target: document.body,
	props: {
		api: API,
		data: fetch(API + "/home_data")
	}
});

export default app;