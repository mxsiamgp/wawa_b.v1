db.createCollection('Logs', {
    capped: true,
    size: 1073741824 //bytes
});
