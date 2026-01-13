import { execSync } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default function globalSetup() {
	const sampleDir = path.resolve(__dirname, '../../sample');
	console.log('Building sample PDF...');
	execSync('make', { cwd: sampleDir, stdio: 'inherit' });
}
