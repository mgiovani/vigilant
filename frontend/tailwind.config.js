export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        dark: {
          950: '#0a0a0a',
          900: '#1a1a1a',
          800: '#2a2a2a',
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
