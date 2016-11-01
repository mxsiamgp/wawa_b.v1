db.createCollection('Orders');
db.Orders.ensureIndex({
    userId: 1,
    createdTime: -1
});
