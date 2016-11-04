db.createCollection('Merchants');
db.Merchants.ensureIndex({
    name: 1
}, {
    unique: true
});
