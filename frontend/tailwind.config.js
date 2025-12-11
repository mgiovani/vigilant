export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        dark: {
          950: '#000000',
          900: '#0a0a0a',
          800: '#171717',
          700: '#262626',
        },
        accent: {
          DEFAULT: '#ffffff',
          hover: '#e5e5e5',
          400: '#a3a3a3',
        },
      },
    },
  },
  plugins: [],
};
