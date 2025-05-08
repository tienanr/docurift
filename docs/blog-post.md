# DocuRift: Breathe New Life into Your Legacy API Documentation

Hey there, fellow developer! 👋

Let me guess - you've inherited a legacy API service. The code works, but the documentation? Well, let's just say it's either non-existent, outdated, or scattered across various Confluence pages and Slack messages. Sound familiar?

We've all been there. That moment when you need to make a change to an API endpoint, but you're not sure what it does, what it expects, or what it returns. You're left with two options: dig through the code (if you can find it) or ask around the team (if anyone remembers).

## The Legacy API Documentation Nightmare

Remember that time when:
- You spent hours reverse-engineering an API endpoint from the code?
- You had to test every possible input to understand what the API expects?
- You found three different versions of the API documentation, none of them correct?
- You had to explain to a new team member why there's no documentation?

Yeah, we've all been there. And it's not fun.

## Meet DocuRift: Your Legacy API's Best Friend

DocuRift is like having a time machine for your API documentation. It watches your API in action and automatically generates accurate, up-to-date documentation. No more guessing, no more digging through code, no more "I think this is how it works" conversations.

The best part? It works with any HTTP API, no matter how old or complex. Whether you're dealing with a 10-year-old service or a modern microservice, DocuRift can help you understand and document it.

## Why You'll Love It

Remember that time you had to figure out what that legacy endpoint does? With DocuRift, you can just point it at your API and let it do the heavy lifting. It automatically:
- Discovers all your endpoints
- Figures out what data they expect
- Documents what they return
- Shows you real examples of requests and responses

It's like having a technical writer who understands your legacy code better than anyone else.

## How Does It Work?

DocuRift acts as an HTTP proxy that sits between your clients and your API server. It passively observes the traffic flowing through it and builds documentation based on real requests and responses.

Here's a step-by-step example of how we use DocuRift in our development environment:

1. First, set up your config file (config.yaml):
```yaml
proxy:
  port: 8080
  backend_url: "http://your-api:8081"

analyzer:
  port: 8082
  max_examples: 10
```

2. Start your API server (in this case, we're using a sample online store API):
```bash
# Build and run the example API
docker build -t online-store -f examples/online_store/Dockerfile .
docker run -p 8081:8081 online-store
```

3. Start DocuRift:
```bash
docurift
```

4. Make some requests to your API through DocuRift:
```bash
# List products
curl http://localhost:8080/products

# Get a specific product
curl http://localhost:8080/products/1

# Create a new product
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name": "New Product", "price": 99.99}'
```

5. Access your automatically generated documentation at `http://localhost:8082/docs`

That's it! Your legacy API documentation is now automatically generated and maintained. No more digging through code, no more outdated docs.

## Let's Connect

I'd love to hear your thoughts and see how DocuRift can help you. You can find me on:
- GitHub: [@tienanr](https://github.com/tienanr)
- Reddit: [@tienanr](https://www.reddit.com/user/tienanr/)

Give DocuRift a try today. Your future self (and your team) will thank you! 🚀

P.S. If you're still manually documenting legacy APIs, you're working too hard. Let DocuRift handle that for you. 😉 