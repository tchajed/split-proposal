import adapter from '@sveltejs/adapter-static';

const base = process.env.BASE_PATH ?? '';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			// SPA mode: fallback to index.html for client-side routing
			fallback: 'index.html',
			pages: 'build',
			assets: 'build'
		}),
		paths: {
			base
		}
	}
};

export default config;
