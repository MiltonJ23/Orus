# Contributing to Orus

First off, thank you for considering contributing to Orus. We are building the definitive reading experience for the desktop—unifying PDF and EPUB formats with a unique Mecha-Egyptian aesthetic.

Orus adheres to a strong software engineering methodology philosophy. We value clean architecture, comprehensive testing, and strict visual fidelity. This document outlines the engineering process, branching strategy, and pipeline requirements to ensure every commit maintains the project's high quality.

## 1. The Engineering Philosophy

* **Clean Architecture Strictness:** We do not mix layers. 
* **Docs-as-Code:** If you change a feature, you must update the documentation in `docs/` in the same Pull Request.
* **Visual Fidelity:** All UI changes must adhere to the Design System (see `docs/design/visual_language.md`).

## 2. Branching Strategy & Lifecycle

We use a strict promotion workflow to ensure stability. Please do not commit directly to `main`, `stage`, or `dev`.

### The Branches

| Branch | Role | CI/CD Pipeline Requirements |
| --- | --- | --- |
| **`main`** | **Production**. Stable releases only. | Builds Release Binaries (Windows/Mac/Linux). Publishes to GitHub Releases. |
| **`stage`** | **Beta / RC**. The "waiting room". | **Full E2E Suite**. Launches the app in a headless environment, simulates user flows (Open Book -> Read -> Close), and verifies tracking logic. |
| **`dev`** | **Integration**. The bleeding edge. | **Unit Tests & Linting**. Runs `go test ./...` (fast) and `golangci-lint`. strict coverage checks. |
| **`feat/*`** | **Work**. Your feature branch. | Sandbox. Must pass Unit Tests before PR. |

### The Promotion Flow

1. **Development:** You branch off `dev` (e.g., `feat/add-epub-parser`). You open a PR targeting `dev`.
2. **Integration:** Once merged to `dev`, the code is live for fellow developers.
3. **Beta Release (The "Stage" Gate):** Periodically, we merge `dev` into `stage`.
* *Trigger:* This triggers the **Heavy Pipeline** (Integration & UI Tests).
* *Stabilization Phase:* Code sits in `stage` for a mandatory period (e.g., 48-72 hours) to allow beta testers to report issues.


4. **Production Release:** If the Stabilization Phase passes with zero critical bugs, we tag the release (SemVer) and merge `stage` into `main`.

## 3. Pull Request Process

To maintain a clean history and high quality, all PRs must adhere to these rules:

### A. Semantic Commits

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. This allows us to automatically generate changelogs.

* `feat: add pdfium rendering engine`
* `fix: resolve memory leak in library scan`
* `docs: update architectural decision records`
* `style: adjust scarab-beetle progress bar color`

### B. The Checklist

Before submitting your PR, ensure:

1. [ ] You have added Unit Tests for any new business logic.
2. [ ] You have run `make fmt` and `make lint` locally.
3. [ ] Your code respects the `internal/` isolation rules.
4. [ ] You have updated `docs/` if you changed architecture or UI scaling.

### C. Visual Changes

If your PR alters the UI (The "Glassy/Mecha" look):

* You **must** include a screenshot or GIF in the PR description.
* Verify that your changes respect the theme constants in `internal/adapter/ui/theme`.


## 4. Reporting Issues

* **Bugs:** Please use the **Bug Report** template. Include your OS, screen resolution, and the file format you were trying to read.
* **Visual Glitches:** Since we use a custom rendering engine, please attach a screenshot of the glitch.
* **Feature Requests:** Use the **RFC (Request for Comments)** template. Discuss the "Why" before the "How".

---

**"May your reading be swift and your knowledge eternal."**