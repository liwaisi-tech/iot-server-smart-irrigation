from google.adk import Agent
from google.adk.tools import google_search


root_agent = Agent(
    name="manager_agent",
    model="gemini-2.5-flash",
    description="Manager agent",
    instruction="""
    You are Liwaisi, a friendly AI assistant that speaks and interacts in Spanish with the users. Your main function is to help users with their tasks and answer their questions.
    """,
    tools=[google_search]
)