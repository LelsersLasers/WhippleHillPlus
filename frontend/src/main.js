import App from './App.svelte';

const API = "http://64.98.192.13:3001";
const app = new App({
	target: document.body,
	props: {
		api: API,
		data: fetch(API + "/home_data")
	}
});

export default app;