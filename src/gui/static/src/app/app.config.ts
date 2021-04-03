export const AppConfig = {
  otcEnabled: false,
  maxHardwareWalletAddresses: 1,
  urlForHwWalletVersionChecking: 'https://version.skycoin.com/skywallet/version.txt',
  hwWalletDownloadUrlAndPrefix: 'https://downloads.skycoin.com/skywallet/skywallet-firmware-v',
  hwWalletDaemonDownloadUrl: 'https://www.skycoin.com/downloads/',

  urlForVersionChecking: 'https://version.skycoin.com/skycoin/version.txt',
  walletDownloadUrl: 'https://www.skycoin.com/downloads/',

  priceApiId: 'sky-skycoin',

  languages: [{
      code: 'en',
      name: 'English',
      iconName: 'en.png',
    },
    {
      code: 'zh',
      name: '中文',
      iconName: 'zh.png',
    },
    {
      code: 'es',
      name: 'Español',
      iconName: 'es.png',
    },
  ],
  defaultLanguage: 'en',

  mediumModalWidth: '566px',
};
