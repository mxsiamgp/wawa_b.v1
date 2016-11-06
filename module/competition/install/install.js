db.createCollection('Competitions');
db.Competitions.ensureIndex({
    name: 1
}, {
    unique: true
});
