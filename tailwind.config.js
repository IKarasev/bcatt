/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [".\\views\\*.{templ,js}"],
  theme: {
    extend: {},
  },
  daisyui: {
    themes: [
      {
        mytheme: {
          primary: "#60a5fa",

          secondary: "#9ca3af",

          accent: "#00ffff",

          neutral: "#f3f4f6",

          "base-100": "#f3f4f6",

          info: "#9ca3af",

          success: "#84cc16",

          warning: "#facc15",

          error: "#f87171",
        },
      },
    ],
  },
  plugins: [require("daisyui")],
};
