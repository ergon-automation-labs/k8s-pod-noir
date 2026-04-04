# Setup & installation — from zero to first case

This guide is for a **fresh machine**: no Docker, no existing Rancher/Desktop Kubernetes, no repo clone yet. At the end you will have a **local cluster**, **`kubectl`** pointed at it, **Docker** running the game container, and **`make run`** opening the file cabinet.

If you already have Docker + a working context, skip to **[§5 Clone and run](#5-clone-the-repo-build-run)**.

---

## 1. Install Docker (engine + Compose)

POD noir’s supported path is **`docker compose`** (Compose V2 plugin), not legacy `docker-compose` as a separate binary.

- **macOS / Windows:** install **[Docker Desktop](https://docs.docker.com/desktop/)** (includes Docker Engine and Compose).
- **Linux:** install **[Docker Engine](https://docs.docker.com/engine/install/)** and the **[Compose plugin](https://docs.docker.com/compose/install/linux/)**.

Verify:

```bash
docker version
docker compose version
```

---

## 2. Install `kubectl`

Install the **kubectl** CLI for your OS ([Kubernetes install tools](https://kubernetes.io/docs/tasks/tools/)).

Verify:

```bash
kubectl version --client
```

---

## 3. Create a local Kubernetes cluster

You need **any** cluster where you can create namespaces and apply workloads. Two common options:

### Option A — kind (recommended; close to CI)

1. Install **[kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)**.
2. Create a cluster:

   ```bash
   kind create cluster --name pod-noir
   ```

3. Point your shell at it (kind usually updates `kubectl` context automatically):

   ```bash
   kubectl cluster-info
   kubectl get nodes
   ```

### Option B — minikube, k3d, or another distro

Follow that product’s “quick start” so **`kubectl get nodes`** works.

### Option C — Rancher Desktop (all-in-one)

Install **[Rancher Desktop](https://docs.rancherdesktop.io/)**; it bundles Docker-compatible tooling and a Kubernetes distribution. Enable Kubernetes in the app, wait until it is ready, then **`kubectl cluster-info`**.

---

## 4. Point the API at Docker (when the API is on `localhost`)

If your kubeconfig uses **`https://127.0.0.1:6443`** (or `localhost`) for the cluster API, **containers** cannot reach that URL as “the host” unless you rewrite it.

**Recommended:** copy your context to a dedicated name and set the server to **`https://host.docker.internal:6443`** (Docker Desktop / Linux with host-gateway), or use the gateway IP your setup documents.

Then:

```bash
kubectl config use-context <that-context>
kubectl cluster-info
```

POD noir’s **`doctor`** command (and **`docker-compose.yml`**) can also help with TLS/rewrite flags — see **`README.md`** and run **`make smoke`** after the first build.

---

## 5. Clone the repo, build, run

```bash
git clone https://github.com/ergon-automation-labs/k8s-pod-noir.git
cd k8s-pod-noir
```

(Use your fork URL if you forked the repo.)

Optional: copy **`.env.example`** to **`.env`** and set **`POD_NOIR_*`** (e.g. LLM keys). Compose loads **`.env`** automatically — see **`README.md`** → Environment.

Build the runtime image and prove the container can reach the API:

```bash
make build
make smoke
```

Start the game (file cabinet → pick a case → REPL):

```bash
make run
```

Or jump straight into a scenario:

```bash
make run RUN_EXTRA='-scenario case-001-overnight-shift'
```

---

## 6. Checks if something fails

| Symptom | What to try |
|--------|-------------|
| **`make smoke`** fails / `doctor` cannot reach API | Fix kubeconfig server URL for Docker (**§4**), confirm **`kubectl cluster-info`** on the **host** works. |
| **Permission / RBAC** errors applying scenarios | Use a context with permission to create namespaces and namespaced resources (kind/minikube default is usually fine). |
| **No Docker** but you have **Go 1.23+** | **`make build-native`** then **`./bin/podnoir doctor`** / **`./bin/podnoir`** on the host — you still need **`kubectl`** + a cluster (**§2–3**). |

---

## 7. Optional: quality-of-life

- **`make playtest-smoke`** — doctor + optional host `kubectl` checks (see **`docs/playtest-checklist.md`**).
- **`make git-hooks`** — install tracked git hooks (`githooks/`) for optional playtest smoke on commit/push.
- **`make test`** — unit tests in Docker (no host Go required).

For contributor workflows, see **[AGENTS.md](../AGENTS.md)**.
