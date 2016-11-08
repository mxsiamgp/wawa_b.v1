db.createCollection('Competitions');
db.Competitions.ensureIndex({
    name: 1
}, {
    unique: true
});

db.createCollection('DrawnTickets');
db.DrawnTickets.ensureIndex({
    orderId: 1,
    orderItemId: 1
}, {
    unique: true
});
db.DrawnTickets.ensureIndex({
    competitionId: 1,
    userId: 1
}, {
    unique: false
});
