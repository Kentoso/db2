db.books.insertMany([
    {
        title: "The Last Chronicle",
        author: { name: "Jane Doe", biography: "An acclaimed author known for dystopian themes." },
        short_description: "A captivating tale of mystery and adventure in a dystopian future.",
        published_date: new Date("2022-01-01"),
        genres: ["Dystopian"]
    },
    {
        title: "Beyond the Horizon",
        author: { name: "John Smith", biography: "Bestselling author of adventure novels." },
        short_description: "An epic journey of discovery and survival in uncharted territories.",
        published_date: new Date("2022-01-15"),
        genres: ["Adventure"]
    },
    {
        title: "Whispers of the Past",
        author: { name: "Emily Johnson", biography: "Historian and writer with a passion for Victorian history." },
        short_description: "A historical novel set in the Victorian era, unraveling family secrets.",
        published_date: new Date("2021-05-10"),
        genres: ["Historical Fiction"]
    },
    {
        title: "Shadows in the Mirror",
        author: { name: "Michael Brown", biography: "A master of psychological thrillers and suspense." },
        short_description: "A psychological thriller about identity and deception.",
        published_date: new Date("2021-10-23"),
        genres: ["Thriller"]
    },
    {
        title: "Echoes of Time",
        author: { name: "Sarah Wilson", biography: "Writes time-travel stories blending science fiction with romance." },
        short_description: "A time-travel story blending romance, history, and science fiction.",
        published_date: new Date("2022-03-08"),
        genres: ["Science Fiction"]
    },
    {
        title: "The Invisible Crown",
        author: { name: "William Green", biography: "A fantasy and political drama novelist." },
        short_description: "A political drama set in a fantasy kingdom, exploring power and morality.",
        published_date: new Date("2022-02-17"),
        genres: ["Fantasy"]
    },
    {
        title: "Rise of the Phoenix",
        author: { name: "Laura Martinez", biography: "Author of war stories and tales of resilience." },
        short_description: "An inspiring story of resilience and rebirth set against a backdrop of war.",
        published_date: new Date("2020-12-12"),
        genres: ["War"]
    },
    {
        title: "The Forgotten Garden",
        author: { name: "David Garcia", biography: "A magical realism author who brings gardens to life." },
        short_description: "A magical realism novel about a mysterious garden holding centuries of secrets.",
        published_date: new Date("2022-04-20"),
        genres: ["Magical Realism"]
    },
    {
        title: "Ocean's Whisper",
        author: { name: "Linda White", biography: "Marine biologist turned novelist, writing about the sea." },
        short_description: "A marine adventure that uncovers the mysteries and beauties of the deep sea.",
        published_date: new Date("2021-07-30"),
        genres: ["Marine Adventure"]
    },
    {
        title: "Stars Beyond Reach",
        author: { name: "Richard Lee", biography: "Science fiction author exploring the cosmos." },
        short_description: "A space opera that explores the farthest reaches of the universe and the human spirit.",
        published_date: new Date("2021-11-11"),
        genres: ["Space Opera"]
    }
]);