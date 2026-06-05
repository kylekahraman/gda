import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://kylekahraman.github.io',
  base: '/gda',
  integrations: [
    starlight({
      title: 'GDA',
      description: 'Content-addressed data versioning for research datasets',
      favicon: '/favicon.svg',
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/kylekahraman/gda' },
      ],
      editLink: {
        baseUrl: 'https://github.com/kylekahraman/gda/edit/master/docs/',
      },
      sidebar: [
        {
          label: 'Getting Started',
          items: [
            { label: 'Installation', slug: 'installation' },
            { label: 'Quick Start', slug: 'quick-start' },
            { label: 'Overview', slug: 'overview' },
          ],
        },
        {
          label: 'Guides',
          items: [
            { label: 'Adding Data', slug: 'guides/adding-data' },
            { label: 'Snapshots & Checkout', slug: 'guides/snapshots' },
            { label: 'Moving & Removing', slug: 'guides/moving-removing' },
            { label: 'Unlock & Lock', slug: 'guides/unlock-lock' },
            { label: 'Remote Storage', slug: 'guides/remote-storage' },
            { label: 'Housekeeping', slug: 'guides/housekeeping' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'Commands', slug: 'reference/commands' },
            { label: 'Store Format', slug: 'reference/store-format' },
            { label: 'Configuration', slug: 'reference/configuration' },
          ],
        },
      ],
    }),
  ],
});
