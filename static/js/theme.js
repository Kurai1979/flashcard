// Theme switching. Loaded synchronously from <head> so the class is on <html>
// before first paint — a deferred script would flash the light theme first.
//
// The choice lives in localStorage, not the session, so it applies to the login
// page too and never needs a round trip. With no stored choice we follow the OS
// setting; picking a theme explicitly pins it.
(function () {
	var root = document.documentElement;

	function stored() {
		try {
			return localStorage.getItem("theme");
		} catch (e) {
			// Private mode or storage disabled: fall back to the OS setting.
			return null;
		}
	}

	function apply(theme) {
		root.classList.toggle("dark", theme === "dark");
	}

	function preferred() {
		return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
	}

	apply(stored() || preferred());

	// Only follow the OS while the user hasn't chosen for themselves.
	window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", function (e) {
		if (!stored()) {
			apply(e.matches ? "dark" : "light");
		}
	});

	// Called from the toggle button's onclick.
	window.toggleTheme = function () {
		var next = root.classList.contains("dark") ? "light" : "dark";
		apply(next);
		try {
			localStorage.setItem("theme", next);
		} catch (e) {
			// Not persisting is survivable; the page still switches.
		}
	};
})();
