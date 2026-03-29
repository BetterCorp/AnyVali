document.documentElement.classList.add("js");

if (window.hljs) {
  window.hljs.highlightAll();
}

const migrationDemo = document.querySelector("[data-migration-demo]");

if (migrationDemo) {
  const tabButtons = Array.from(migrationDemo.querySelectorAll("[data-lang-tab]"));
  const langLabel = migrationDemo.querySelector("[data-lang-label]");
  const tabScroller = migrationDemo.querySelector(".migration-tabs");
  const scrollButtons = Array.from(migrationDemo.querySelectorAll("[data-scroll-dir]"));

  const labels = {
    typescript: "TypeScript",
    python: "Python",
    go: "Go",
    java: "Java",
    csharp: "C#",
    rust: "Rust",
    php: "PHP",
    ruby: "Ruby",
    kotlin: "Kotlin",
    cpp: "C++",
  };

  let active = "python";
  let rotateTimer;

  const updateScrollButtons = () => {
    if (!tabScroller) return;
    const maxScroll = tabScroller.scrollWidth - tabScroller.clientWidth;
    for (const button of scrollButtons) {
      const dir = button.dataset.scrollDir;
      if (dir === "left") {
        button.disabled = tabScroller.scrollLeft <= 4;
      } else {
        button.disabled = tabScroller.scrollLeft >= maxScroll - 4;
      }
    }
  };

  const renderSnippet = (lang) => {
    active = lang;
    migrationDemo.classList.remove("lang-python", "lang-go", "lang-typescript");
    migrationDemo.classList.remove(
      "lang-java",
      "lang-csharp",
      "lang-rust",
      "lang-php",
      "lang-ruby",
      "lang-kotlin",
      "lang-cpp",
    );
    migrationDemo.classList.add(`lang-${lang}`);

    if (langLabel) {
      langLabel.textContent = labels[lang];
    }

    for (const button of tabButtons) {
      button.classList.toggle("is-active", button.dataset.langTab === lang);
    }
  };

  const queueNext = () => {
    window.clearTimeout(rotateTimer);
    rotateTimer = window.setTimeout(() => {
      const order = [
        "typescript",
        "python",
        "go",
        "java",
        "csharp",
        "rust",
        "php",
        "ruby",
        "kotlin",
        "cpp",
      ];
      const currentIndex = order.indexOf(active);
      const nextIndex = (currentIndex + 1) % order.length;
      renderSnippet(order[nextIndex]);
      queueNext();
    }, 2200);
  };

  for (const button of tabButtons) {
    button.addEventListener("click", () => {
      renderSnippet(button.dataset.langTab);
      queueNext();
    });
  }

  for (const button of scrollButtons) {
    button.addEventListener("click", () => {
      if (!tabScroller) return;
      const dir = button.dataset.scrollDir === "left" ? -1 : 1;
      tabScroller.scrollBy({ left: dir * 160, behavior: "smooth" });
    });
  }

  if (tabScroller) {
    tabScroller.addEventListener("scroll", updateScrollButtons);
    window.addEventListener("resize", updateScrollButtons);
  }

  renderSnippet(active);
  updateScrollButtons();
  queueNext();
}
