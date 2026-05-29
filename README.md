# 🐊 Gator

**Gator** is a lightweight, command-line RSS feed aggregator built in Go.

Originally created as part of the [Boot.dev Build a Blog Aggregator course](https://www.boot.dev/courses/build-blog-aggregator-golang), this project serves as a practical exploration of Go project architecture, CLI tooling, and XML parsing.

Gator takes a **SQL-first approach** to database management. By utilizing `sqlc` and `goose`, it leverages raw SQL—the most universal database language—to completely avoid the overhead, custom syntax, and N+1 query problems common with traditional ORMs.

### Features

* **User Management:** Register and log in to personalized accounts.
* **Feed Management:** Add, follow, and unfollow RSS feeds.
* **Background Aggregation:** Fetch posts automatically on a customizable, continuous schedule.
* **Post Browsing:** View the latest curated content directly from your terminal.

---

## ⚙️ Configuration

Before running Gator, you need to create a configuration file named `.gatorconfig.json` in your user's home directory. This file stores your database connection string and tracks the currently logged-in user.

Create `~/.gatorconfig.json` and add your database URL (e.g., PostgreSQL):

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator"
}

```

---

## 🚀 Installation

There are three ways to install Gator on your system:

### 1. Go Install (Recommended)

If you have the Go toolchain installed, you can easily build and install the binary directly into your `$GOPATH/bin` folder:

```bash
go install github.com/edu292/gator@latest

```

### 2. Pre-built Binaries

Navigate to the **Releases** section on the right side of this repository. You can download a pre-compiled binary specific to your operating system (Windows, macOS, or Linux).

### 3. Build from Source

To clone the repository and compile it yourself:

```bash
git clone https://github.com/edu292/gator.git
cd gator

# Run directly:
go run . [command...]

# Or build the executable:
go build -o gator .
./gator [command]

```

---

## 💻 Usage

Once configured and installed, you can use the `gator` CLI to manage your feeds.

*Note: Arguments in `<angle brackets>` are required, while `[square brackets]` are optional.*

### Authentication

| Command | Description |
| --- | --- |
| `gator register <username>` | Creates a new user account and immediately logs you in. |
| `gator login <username>` | Logs in as an existing user. |

### Managing Feeds

| Command | Description |
| --- | --- |
| `gator addfeed <name> <url>` | Creates a new RSS feed in the database and automatically follows it. |
| `gator follow <url>` | Starts following an existing feed. |
| `gator unfollow <url>` | Stops following a specific feed. |

### Fetching & Reading

| Command | Description |
| --- | --- |
| `gator agg <time_interval>` | Starts the aggregator worker to fetch new posts. The interval should be formatted as a Go duration (e.g., `1m`, `30s`). Keep this process running to continuously fetch posts. |
| `gator browse [limit]` | Displays the most recent posts from the feeds you follow. `limit` defaults to 2 if not provided. |

---

**Example Workflow:**

```bash
# 1. Create your account
gator register alice

# 2. Add a feed you want to read
gator addfeed "Boot.dev Blog" https://blog.boot.dev/index.xml

# 3. Start the background worker to fetch posts every minute (leave this running in a separate terminal)
gator agg 1m

# 4. Browse your newly fetched posts (shows the 5 most recent)
gator browse 5

```
