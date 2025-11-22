import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'index',
    'installation',
    'usage',
    'configuration',
    {
      type: 'category',
      label: 'Backends',
      items: ['backends/s3', 'backends/local'],
    },
    {
      type: 'category',
      label: 'Targets',
      items: ['targets/webhook', 'targets/slack', 'targets/stdout'],
    },
    'persistence',
    'metrics',
    'contributing',
  ],
};

export default sidebars;
