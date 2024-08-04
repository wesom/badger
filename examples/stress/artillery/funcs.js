module.exports = { createRandomScore };

function createRandomScore(userContext, events, done) {
  const data = {
    timestamp: Date.now(),
    score: Math.floor(Math.random() * 100)
  };

  userContext.vars.data = data;

  return done();
}