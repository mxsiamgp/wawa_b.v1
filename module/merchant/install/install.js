db.createCollection('Merchants');
db.Merchants.ensureIndex({
    name: 1
}, {
    unique: true
});
db.Merchants.ensureIndex({
    managerUserId: 1
}, {
    unique: true
});
