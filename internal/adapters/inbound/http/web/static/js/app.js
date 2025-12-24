(() => {
  const storedTheme = localStorage.getItem("theme");
  if (storedTheme === "dark") {
    document.documentElement.classList.remove("light");
    document.documentElement.classList.add("dark");
  }

  const toggle = document.getElementById("theme-toggle");
  if (toggle) {
    toggle.addEventListener("click", () => {
      const root = document.documentElement;
      const isDark = root.classList.contains("dark");
      if (isDark) {
        root.classList.remove("dark");
        root.classList.add("light");
        localStorage.setItem("theme", "light");
        return;
      }
      root.classList.remove("light");
      root.classList.add("dark");
      localStorage.setItem("theme", "dark");
    });
  }

  let lastOk = true;
  setInterval(() => {
    fetch("/health", { cache: "no-store" })
      .then(() => {
        if (!lastOk) {
          location.reload();
        }
        lastOk = true;
      })
      .catch(() => {
        lastOk = false;
      });
  }, 1000);
})();
