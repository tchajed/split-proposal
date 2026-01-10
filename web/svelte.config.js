import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		adapter: adapter({
			// SPA mode: fallback to index.html for client-side routing
			fallback: 'index.html',
			pages: 'build',
			assets: 'build'
		})
	}
};

export default config;
