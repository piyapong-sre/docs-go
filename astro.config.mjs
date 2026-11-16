// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	integrations: [
		starlight({
			title: 'Go ภาษาไทย',
			description: 'เรียนภาษา Go ตั้งแต่พื้นฐานจนถึงขั้นสูง และสร้าง Web API แบบ best practice',
			defaultLocale: 'root',
			locales: {
				root: { label: 'ไทย', lang: 'th' },
			},
			customCss: [
				'@fontsource/ibm-plex-sans-thai/400.css',
				'@fontsource/ibm-plex-sans-thai/600.css',
				'./src/styles/custom.css',
			],
			lastUpdated: false,
			pagination: true,
			tableOfContents: { minHeadingLevel: 2, maxHeadingLevel: 3 },
			expressiveCode: {
				themes: ['github-dark', 'github-light'],
				styleOverrides: { borderRadius: '0.5rem' },
			},
			sidebar: [
				{ label: '1. เริ่มต้น', items: [{ autogenerate: { directory: 'start' } }] },
				{ label: '2. พื้นฐาน', items: [{ autogenerate: { directory: 'basics' } }] },
				{ label: '3. ระดับกลาง', items: [{ autogenerate: { directory: 'intermediate' } }] },
				{ label: '4. Concurrency', items: [{ autogenerate: { directory: 'concurrency' } }] },
				{ label: '5. ขั้นสูง', items: [{ autogenerate: { directory: 'advanced' } }] },
				{ label: '6. สร้าง Web API', items: [{ autogenerate: { directory: 'web-api' } }] },
			],
		}),
	],
});
