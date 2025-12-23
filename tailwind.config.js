/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/adapters/inbound/http/web/templates/**/*.{html,tmpl,gohtml}",
    "./internal/adapters/inbound/http/web/handlers/**/*.go",
    "./internal/adapters/inbound/http/web/viewmodels/**/*.go",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
};
