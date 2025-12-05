import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'index',
    'installation',
    'usage',
    'configuration',
    {
      type: 'category',
      label: 'Targets',
      items: ['targets/webhook', 'targets/slack'],
    },
    'contributing',
  ],
};

export default sidebars;
